package utils

import (
	"strings"
)

func SafeGetString(data map[string]interface{}, key string) string {
	if value, ok := data[key]; ok && value != nil {
		return value.(string)
	}
	return ""
}

func IsValidEthereumAddress(address string) bool {
	if len(address) != 42 {
		return false
	}
	if !strings.HasPrefix(address, "0x") {
		return false
	}
	for _, char := range address[2:] {
		if (char < '0' || char > '9') && (char < 'a' || char > 'f') && (char < 'A' || char > 'F') {
			return false
		}
	}
	return true
}
