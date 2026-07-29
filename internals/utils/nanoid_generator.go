package utils

import (
	"crypto/rand"
	"encoding/hex"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"

// GenerateNanoId creates a short random identifier for request tracking and job naming.
func GenerateNanoId(length int) string {
	id, err := gonanoid.Generate(alphabet, length)
	if err == nil {
		return id
	}

	// Fall back to a random hex string if the primary generator fails.
	buffer := make([]byte, (length+1)/2)
	_, _ = rand.Read(buffer)
	s := hex.EncodeToString(buffer)
	s = s[:length]

	return s
}
