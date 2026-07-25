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
	buf := make([]byte, (length+1)/2)
	_, _ = rand.Read(buf)
	s := hex.EncodeToString(buf)

	if len(s) > length {
		s = s[:length]
	}
	return s
}
