#!/bin/bash

GREEN='\033[1;32m'
PURPLE='\033[1;35m'
NC='\033[0m'

step() {
  echo -e "\n${GREEN}[$1/8] $2${NC}"
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
    if [ $? -ne 0 ]; then
        echo "Failed to create user. Installation aborted."
        exit 1
    fi
fi

# Download the appropriate archive
TEMP_DIR=$(mktemp -d)
ARCHIVE_NAME="obfs-exporter-${VERSION}-linux-${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/hteppl/obfs-exporter/releases/download/${VERSION}/${ARCHIVE_NAME}"

step 2 "Downloading binary from: ${DOWNLOAD_URL}"
curl -L -o "${TEMP_DIR}/${ARCHIVE_NAME}" "${DOWNLOAD_URL}"
if [ $? -ne 0 ]; then
    echo "Failed to download binary. Installation aborted."
    rm -rf "${TEMP_DIR}"
    exit 1
fi

# Extract binary
step 3 "Extracting binary..."
tar -xzf "${TEMP_DIR}/${ARCHIVE_NAME}" -C "${TEMP_DIR}"
if [ $? -ne 0 ]; then
    echo "Failed to extract binary. Installation aborted."
    rm -rf "${TEMP_DIR}"
    exit 1
fi

# Force remove old binary if exists
if [ -f /usr/local/bin/obfs-exporter ]; then
    rm -f /usr/local/bin/obfs-exporter
    if [ $? -ne 0 ]; then
        echo "Failed to remove old binary. Installation aborted."
        rm -rf "${TEMP_DIR}"
        exit 1
    fi
fi

# Install binary to /usr/local/bin
step 4 "Installing binary to /usr/local/bin..."
cp "${TEMP_DIR}/obfs-exporter" /usr/local/bin/
if [ $? -ne 0 ]; then
    echo "Failed to install binary. Installation aborted."
    rm -rf "${TEMP_DIR}"
    exit 1
fi

# Clean up and set permissions
rm -rf "${TEMP_DIR}"
chmod 755 /usr/local/bin/obfs-exporter

# Create config directory
step 5 "Creating configuration directory..."
mkdir -p /etc/obfs-exporter/
if [ $? -ne 0 ]; then
    echo "Failed to create config directory. Installation aborted."
    exit 1
fi

# Check if config file already exists
CONFIG_FILE="/etc/obfs-exporter/config.yaml"
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
    echo "Downloading example config from GitHub..."
    curl -s -o "$CONFIG_FILE" https://raw.githubusercontent.com/hteppl/obfs-exporter/main/config-example.yaml
    if [ $? -ne 0 ]; then
        echo "Failed to download config file. Installation aborted."
        exit 1
    fi

    # Interactive configuration
    echo "Provide your 3X-UI panel details:"

    # Get Panel Port
    while true; do
        read -p "Enter panel port (e.g., 2053): " PANEL_PORT
        if [[ "$PANEL_PORT" =~ ^[0-9]+$ ]] && [ "$PANEL_PORT" -ge 1 ] && [ "$PANEL_PORT" -le 65535 ]; then
            break
        else
            echo "Error: Port must be a number between 1 and 65535."
        fi
    done

    # Get Panel Path
    while true; do
        read -p "Enter panel path (e.g., /panel/ or /xui/, but usually /ADKhilALhvlzkjcBLHJXCEp/ or something similar): " PANEL_PATH
        # Normalise: remove leading and trailing slashes
        PANEL_PATH=$(echo "$PANEL_PATH" | sed 's:^/*::; s:/*$::')
        if [ -z "$PANEL_PATH" ]; then
            echo "Error: Panel path cannot be empty."
        else
            break
        fi
    done

    # Build the full base URL
    PANEL_URL="http://localhost:${PANEL_PORT}/${PANEL_PATH}"

    # Get credentials
    while true; do
        read -p "Enter Panel Username: " PANEL_USERNAME
        if [ -z "$PANEL_USERNAME" ]; then
            echo "Error: Panel Username cannot be empty. Please try again."
        else
            break
        fi
    done

    while true; do
        read -s -p "Enter Panel Password: " PANEL_PASSWORD
        echo ""
        if [ -z "$PANEL_PASSWORD" ]; then
            echo "Error: Panel Password cannot be empty. Please try again."
        else
            break
        fi
    done

    # Validate connection to panel
    echo "Validating connection to panel..."
    TEMP_RESPONSE=$(mktemp)
    CURL_EXIT_CODE=0
    LOGIN_RESULT=$(curl -s -w "%{http_code}" -o "$TEMP_RESPONSE" -X POST "${PANEL_URL}/login" \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"${PANEL_USERNAME}\",\"password\":\"${PANEL_PASSWORD}\"}" || { CURL_EXIT_CODE=$?; echo "000"; })

    if [ $CURL_EXIT_CODE -ne 0 ]; then
        echo "Failed to connect to panel. Network error (curl exit code: $CURL_EXIT_CODE)"
        echo "Please check if the panel URL is correct and the server is reachable."
        read -p "Continue anyway? (y/n): " CONTINUE
        if [ "$CONTINUE" != "y" ] && [ "$CONTINUE" != "Y" ]; then
            rm -f "$TEMP_RESPONSE"
            echo "Installation aborted."
            exit 1
        fi
    elif [ "$LOGIN_RESULT" == "200" ] || [ "$LOGIN_RESULT" == "303" ]; then
        echo "Successfully connected to panel!"
    elif [ "$LOGIN_RESULT" == "401" ] || [ "$LOGIN_RESULT" == "403" ]; then
        echo "Authentication failed. Invalid username or password."
        read -p "Continue anyway? (y/n): " CONTINUE
        if [ "$CONTINUE" != "y" ] && [ "$CONTINUE" != "Y" ]; then
            rm -f "$TEMP_RESPONSE"
            echo "Installation aborted."
            exit 1
        fi
    else
        echo "Failed to connect to panel. HTTP status: ${LOGIN_RESULT}"
        echo "Please verify your panel URL and credentials."
        read -p "Continue anyway? (y/n): " CONTINUE
        if [ "$CONTINUE" != "y" ] && [ "$CONTINUE" != "Y" ]; then
            rm -f "$TEMP_RESPONSE"
            echo "Installation aborted."
            exit 1
        fi
    fi
    rm -f "$TEMP_RESPONSE"

    # Update the config file with user input
    echo "Updating configuration file with provided details..."
    # Escape special characters in variables for sed
    PANEL_PATH_ESCAPED=$(echo "$PANEL_PATH" | sed 's/[\\/"]/\\&/g')
    PANEL_USERNAME_ESCAPED=$(echo "$PANEL_USERNAME" | sed 's/[\/&]/\\&/g')
    PANEL_PASSWORD_ESCAPED=$(echo "$PANEL_PASSWORD" | sed 's/[\/&]/\\&/g')

    sed -i "s|^\s*panel-username:.*|    panel-username: \"${PANEL_USERNAME_ESCAPED}\"|" "$CONFIG_FILE"
    sed -i "s|^\s*panel-password:.*|    panel-password: \"${PANEL_PASSWORD_ESCAPED}\"|" "$CONFIG_FILE"
    sed -i "s|^\s*panel-port:.*|    panel-port: ${PANEL_PORT}|" "$CONFIG_FILE"
    sed -i "s|^\s*panel-path:.*|    panel-path: \"${PANEL_PATH_ESCAPED}\"|" "$CONFIG_FILE"


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
if [ $? -ne 0 ]; then
    echo "Failed to reload systemd. Installation aborted."
    exit 1
fi

# Enable and start (or restart) the service
step 8 "Enabling and starting obfs-exporter service..."
if systemctl is-active --quiet obfs-exporter.service; then
    echo "Service is already running. Restarting..."
    systemctl restart obfs-exporter.service
    if [ $? -ne 0 ]; then
        echo "Failed to restart service. Installation aborted."
        exit 1
    fi
else
    systemctl enable obfs-exporter.service
    if [ $? -ne 0 ]; then
        echo "Failed to enable service. Installation aborted."
        exit 1
    fi

    systemctl start obfs-exporter.service
    if [ $? -ne 0 ]; then
        echo "Failed to start service. Installation aborted."
        exit 1
    fi
fi

sudo systemctl status obfs-exporter --no-pager

echo -e "\n${PURPLE}✅ 3X-UI Exporter is installed!"
echo -e "${GREEN}\nCheck status:      ${NC}sudo systemctl status obfs-exporter --no-pager"
echo -e "${GREEN}Binary path:       ${NC}/usr/local/bin/obfs-exporter"
echo -e "${GREEN}Config path:       ${NC}$CONFIG_FILE"
echo ""
echo -e "You can view logs with: journalctl -u obfs-exporter.service"
echo -e "Support the project: \033[1;33mhttps://pay.cloudtips.ru/p/67507843${NC}"
echo ""
