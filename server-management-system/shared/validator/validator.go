package validator

import (
	"net"
	"regexp"
	"strings"
)

var (
	// IPv4Regex validates IPv4 format
	ipv4Regex = regexp.MustCompile(`^(\d{1,3})\.(\d{1,3})\.(\d{1,3})\.(\d{1,3})$`)
)

// IsValidIPv4 checks if the given string is a valid IPv4 address
func IsValidIPv4(ip string) bool {
	if net.ParseIP(ip) == nil {
		return false
	}
	// Also check regex for strict IPv4 (no IPv6)
	return ipv4Regex.MatchString(ip)
}

// IsValidEmail performs basic email format validation
func IsValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return false
	}
	return len(email) <= 255
}

// IsValidServerID checks server_id format (alphanumeric, hyphens, underscores)
func IsValidServerID(id string) bool {
	if len(id) < 3 || len(id) > 100 {
		return false
	}
	matched, _ := regexp.MatchString(`^[A-Za-z0-9\-_]+$`, id)
	return matched
}

// TrimAndLower trims whitespace and converts to lowercase
func TrimAndLower(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// IsValidPort checks if port is in valid range
func IsValidPort(port int) bool {
	return port > 0 && port <= 65535
}
