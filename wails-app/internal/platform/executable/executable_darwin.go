//go:build darwin

package executable

import "fmt"

// GetPublisherName returns the organization name from the code signature (Not Implemented).
func GetPublisherName(filePath string) (string, error) {
	return "", fmt.Errorf("not implemented on darwin")
}

// GetProductName returns the product name (Not Implemented).
func GetProductName(exePath string) (string, error) {
	return "", fmt.Errorf("not implemented on darwin")
}

// IsMicrosoftSigned checks if the executable is signed by Microsoft (Not Implemented).
func IsMicrosoftSigned(exePath string) bool {
	return false
}
