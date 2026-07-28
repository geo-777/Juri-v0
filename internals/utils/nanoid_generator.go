package utils

import (
	"crypto/rand"
	"encoding/hex"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"

func GenerateNanoId(length int) string {
	id, err := gonanoid.Generate(alphabet, length)
	if err == nil {
		return id
	}

	// Fallback to a random hex string in case nanoid fails
	buffer := make([]byte, (length+1)/2)
	_, _ = rand.Read(buffer)
	s := hex.EncodeToString(buffer)
	s = s[:length]

	return s
}
