package util

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

func GenerateToken() (token string, hash [32]byte, err error) {
	token, err = RandomToken()
	if err != nil {
		return "", [32]byte{}, err
	}
	hash = sha256.Sum256([]byte(token))

	return token, hash, nil
}

func RandomToken() (string, error) {
	var randomBytes [32]byte
	if _, err := rand.Read(randomBytes[:]); err != nil {
		return "", err
	}

	token := base64.URLEncoding.EncodeToString(randomBytes[:])

	return token, nil
}
