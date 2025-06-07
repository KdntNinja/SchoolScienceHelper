package utils

import (
	"crypto/rand"
	"fmt"
	"time"
)

// GeneratePublicID creates a random public ID for sharing projects.
func GeneratePublicID() string {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		return fmt.Sprintf("pub_%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("pub_%x", b)
}
