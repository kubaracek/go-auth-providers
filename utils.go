package lib

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"strings"
)

// RandomToken generates a random characters of the given length using base64 encoding (safe for URLs).
// Param Length is bound between 32-64 characters to ensure a minimum level of security.
func RandomToken(len int8) (string, error) {
	if len < 32 {
		len = 32
	} else if len > 64 {
		len = 64
	}
	b := make([]byte, 48)
	// if rand fails, we have bigger problems
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		return "", err
	}

	return strings.TrimRight(base64.URLEncoding.EncodeToString(b), "=")[:len], nil
}

// MapOfStrings is a map of strings that implements Params interface
type MapOfStrings map[string]string

func (m MapOfStrings) Get(key string) string {
	return m[key]
}
