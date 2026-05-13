package client

import (
	"crypto/sha512"
	"encoding/base64"
)

func DeviceID(accountIdentifier string) string {
	hash := sha512.Sum512([]byte(accountIdentifier))
	return base64.StdEncoding.EncodeToString(hash[:])
}
