package Hasher

import (
	"crypto/sha512"
	"encoding/base64"
)

// EncodeSha512Base64 Create a SHA512 and Base64 coded hash from a string
func EncodeSha512Base64(passwd string) string {
	sha512 := sha512.Sum512([]byte(passwd))
	encoded := base64.StdEncoding.EncodeToString(sha512[:])
	return encoded
}
