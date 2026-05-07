#!/bin/bash

GREEN='\033[1;32m'
PURPLE='\033[1;35m'
NC='\033[0m'

step() {
  echo -e "\n${GREEN}[$1/8] $2${NC}"
}

# ------------------------------------------------------------
# abort_on_error <error_message> [folder_to_remove]
#
# Must be called immediately after the command you want to guard.
# Arguments:
#   $1 (optional) – Error message to print on failure.
#   $2 (optional) – Directory to remove (with `rm -rf`) on failure.
#
# If the previous command failed (exit ≠ 0), the function:
#   1. Prints the provided error message to stderr, or a default one, if none was provided.
#   2. If $2 is given and not empty, deletes that directory.
#   3. Exits the script with the same non‑zero exit code.
# ------------------------------------------------------------
abort_on_error() {
    local last_exit=$?
    if [[ $last_exit -ne 0 ]]; then
        local error_message="${1:-}"
        if [[ -z $error_message ]]; then
            error_message="Something went wrong during installation, exiting."
        fi
        echo "$error_message" >&2
        if [[ -n "${2:-}" ]]; then
            echo "Cleaning up directory: $2" >&2
            rm -rf "$2"
        fi
        exit "$last_exit"
    fi
}

# ------------------------------------------------------------------
# prompt_input <var_name> <prompt_text> <default_val> [validation] [--secret]
#
# Prompts the user repeatedly until a valid answer is given.
#   validation  – optional: "number", "nonempty", "port", "bool", … or a custom function name.
#   --secret    – optional: hides typed characters (useful for passwords).
#
# When --secret is used, the prompt shows "[hidden]" instead of the default.
# Pressing Enter still accepts the default value (even if hidden).
# ------------------------------------------------------------------
prompt_input() {
    local var_name="$1"
    local prompt_text="$2"
    local default_val="$3"

    local validation=""
    local secret=0

    shift 3
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --secret|--password)
                secret=1
                shift
                ;;
            *)
                if [[ -z "$validation" ]]; then
                    validation="$1"
                else
                    echo "prompt_input: unexpected argument '$1'" >&2
                    return 1
                fi
                shift
                ;;
        esac
    done

    local input_val
    while true; do
        if [[ -z "$default_val" ]]; then
                read -p "${prompt_text} (required): " input_val
            else
                read -p "${prompt_text} [${default_val}]: " input_val
            fi
        fi

        if [[ -z "$default_val" && -z "$input_val" ]]; then
            echo "  This value is required." >&2
            continue
        fi
        input_val="${input_val:-$default_val}"

        # Validation logic
        case "$validation" in
            number)
                if [[ "$input_val" =~ ^[0-9]+$ ]]; then break
                else echo "  Please enter a positive integer." >&2; fi
                ;;
            nonempty)
                if [[ -n "$input_val" ]]; then break
                else echo "  This value cannot be empty." >&2; fi
                ;;
            port)
                if [[ "$input_val" =~ ^[0-9]+$ ]] && (( input_val >= 1 && input_val <= 65535 )); then break
                else echo "  Please enter a valid port (1-65535)." >&2; fi
                ;;
            bool|boolean)
                case "${input_val,,}" in
                    true|yes|1)   input_val="true";  break ;;
                    false|no|0)   input_val="false"; break ;;
                    *)            echo "  Please answer 'true' or 'false'." >&2 ;;
                esac
                ;;
            "") break ;;   # no validation → accept anything
            *)
                # Assume it's a custom function name
                if declare -F "$validation" &>/dev/null && "$validation" "$input_val"; then
                    break
                else
                    echo "  Invalid input. Please try again." >&2
                fi
                ;;
        esac
    done

    export "$var_name"="$input_val"
}


# Check if script is run as root
if [ "$(id -u)" -ne 0 ]; then
    echo "Error: This script must be run as root (sudo)."
    exit 1
fi

# Determine system architecture
ARCH=$(uname -m)
case ${ARCH} in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64)
        ARCH="arm64"
        ;;
    *)
        echo "Unsupported architecture: ${ARCH}"
        exit 1
        ;;
esac

# Get latest release tag

# TODO add option to specify version
echo "Fetching latest release information..."
LATEST_RELEASE=$(curl -s https://api.github.com/repos/hteppl/obfs-exporter/releases/latest)
if [ $? -ne 0 ] || [ -z "$LATEST_RELEASE" ]; then
    echo "Failed to fetch release information. Installation aborted."
    exit 1
fi

VERSION=$(echo "${LATEST_RELEASE}" | grep -Po '"tag_name": "\K.*?(?=")')
echo -e "\n${PURPLE}✨ Starting 3X-UI Exporter $VERSION automated install wizard...\033[0m"

# Create dedicated system user for running the service
step 1 "Creating obfs-exporter user"
if ! id -u obfs-exporter > /dev/null 2>&1; then
    useradd -r -s /bin/false obfs-exporter
    abort_on_error "Failed to create user. Installation aborted."
fi

# Download the appropriate archive
TEMP_DIR=$(mktemp -d)
ARCHIVE_NAME="obfs-exporter-${VERSION}-linux-${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/PlushGuardian/obfs-exporter/releases/download/${VERSION}/${ARCHIVE_NAME}"

step 2 "Downloading binary from: ${DOWNLOAD_URL}"
curl -L -o "${TEMP_DIR}/${ARCHIVE_NAME}" "${DOWNLOAD_URL}"
abort_on_error "Failed to download binary. Installation aborted." "${TEMP_DIR}"


# Extract binary
step 3 "Extracting binary..."
tar -xzf "${TEMP_DIR}/${ARCHIVE_NAME}" -C "${TEMP_DIR}"
abort_on_error "Failed to extract binary. Installation aborted." "${TEMP_DIR}"

# Force remove old binary if exists
if [ -f /usr/local/bin/obfs-exporter ]; then
    rm -f /usr/local/bin/obfs-exporter
    abort_on_error "Failed to remove old binary. Installation aborted." "${TEMP_DIR}"
fi

# Install binary to /usr/local/bin
step 4 "Installing binary to /usr/local/bin..."
cp "${TEMP_DIR}/obfs-exporter" /usr/local/bin/
abort_on_error "Failed to install binary. Installation aborted." "${TEMP_DIR}"


# Clean up and set permissions
rm -rf "${TEMP_DIR}"
chmod 755 /usr/local/bin/obfs-exporter


CONFIG_DIR="/etc/obfs-exporter"
# Create config directory
step 5 "Creating configuration directory..."
mkdir -p ${CONFIG_DIR}
abort_on_error "Failed to create config directory. Installation aborted."

# Check if config file already exists
CONFIG_FILE="${CONFIG_DIR}/config.yaml"  # TODO figure out what to do with an existing configuration
SKIP_CONFIG_SETUP=0
if [ -f "$CONFIG_FILE" ]; then
    echo "Configuration file already exists at $CONFIG_FILE"
    while true; do
        read -p "Do you want to overwrite the existing config? (y/N): " yn
        case $yn in
            [Yy]* )
                echo "Overwriting existing configuration..."
                break
                ;;
            * )
                echo "Skipping config setup."
                SKIP_CONFIG_SETUP=1
                ;;
        esac
        [ $SKIP_CONFIG_SETUP -eq 1 ] && break
    done
fi

if [ $SKIP_CONFIG_SETUP -eq 0 ]; then
    # Download example config file
    echo "Downloading config template from GitHub..."
    curl -s -o "$CONFIG_FILE.tmpl" https://raw.githubusercontent.com/PlushGuardian/obfs-exporter/main/config.yaml.tmpl
    abort_on_error "Failed to download config file template. Installation aborted."

    # Interactive configuration
    echo "General settings:"
    # obfs-exporter section
    prompt_input SCRAPE_TIMEOUT                "obfs-exporter scrape-timeout"            "30"        number
    prompt_input METRICS_PORT                  "obfs-exporter metrics-port"              "9100"      port
    prompt_input METRICS_PATH                  "obfs-exporter metrics-path"              "/metrics"  nonempty
    echo

    # threexui section
    echo "3X-UI settings:"
    prompt_input THREEXUI_PANEL_PORT           "3x-ui panel port"                        "2053"      port
    prompt_input THREEXUI_PANEL_PATH           "3x-ui panel path"                        ""          nonempty
    prompt_input THREEXUI_USERNAME             "3x-ui username"                          ""          nonempty
    prompt_input THREEXUI_PASSWORD             "3x-ui password"                          ""          nonempty --secret
    prompt_input THREEXUI_INSECURE_SKIP_VERIFY "3x-ui insecure skip verify (true|false)" "false"     boolean
    prompt_input THREEXUI_CLIENTS_BYTES_ROWS   "3x-ui clients bytes rows (0=all)"        "0"         number
    prompt_input THREEXUI_TIMEOUT              "3x-ui request timeout (seconds)"         "120"       number
    echo

    # mtproxymax section
    echo "MTProxyMax settings:"
    prompt_input MTPROXYMAX_METRICS_PORT       "MTProxyMax metrics-port"                 "9090"      port
    prompt_input MTPROXYMAX_METRICS_PATH       "MTProxyMax metrics-path"                 "/metrics"  nonempty
    echo

    echo "Generating config file..."

    if command -v envsubst >/dev/null 2>&1; then
        envsubst < $CONFIG_FILE.tmpl > $CONFIG_FILE
        echo "Success! Configuration saved to $CONFIG_FILE."
    else
        echo "Error: 'envsubst' is not installed. (Usually found in gettext package)"
        exit 1
    fi

else
    echo "Using existing configuration file without changes."
fi

chmod 644 "$CONFIG_FILE"
chown -R obfs-exporter:obfs-exporter /etc/obfs-exporter

# Create systemd service file
step 6 "Downloading systemd service file from GitHub..."
curl -s -o /etc/systemd/system/obfs-exporter.service

if [ $? -ne 0 ]; then
    echo "Failed to create service file. Installation aborted."
    exit 1
fi

sed -i "s|^Description=\(.*\)|Description=\1 ${VERSION}|" /etc/systemd/system/obfs-exporter.service
chmod 644 /etc/systemd/system/obfs-exporter.service

# Reload systemd to recognize the new service
step 7 "Reloading systemd daemon..."
systemctl daemon-reload
abort_on_error "Failed to reload systemd. Installation aborted."

# Enable and start (or restart) the service
step 8 "Enabling and starting obfs-exporter service..."
if systemctl is-active --quiet obfs-exporter.service; then
    echo "Service is already running. Restarting..."
    systemctl restart obfs-exporter.service
    abort_on_error "Failed to restart service. Installation aborted."

else
    systemctl enable obfs-exporter.service
    abort_on_error "Failed to enable service. Installation aborted."

    systemctl start obfs-exporter.service
    abort_on_error "Failed to start service. Installation aborted."
fi

sudo systemctl status obfs-exporter --no-pager

echo -e "\n${PURPLE}✅ 3X-UI Exporter is installed!"
echo -e "${GREEN}\nCheck status:      ${NC}sudo systemctl status obfs-exporter --no-pager"
echo -e "${GREEN}Binary path:       ${NC}/usr/local/bin/obfs-exporter"
echo -e "${GREEN}Config path:       ${NC}$CONFIG_FILE"
echo ""
echo -e "You can view logs with: journalctl -u obfs-exporter.service"
echo ""
