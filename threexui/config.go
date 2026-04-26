package threexui

import "time"

// Config holds the connection parameters for a 3X-UI panel.
type Config struct {
	BaseURL            string
	Username           string
	Password           string
	InsecureSkipVerify bool
	ClientsBytesRows   int
	Timeout            time.Duration
}
