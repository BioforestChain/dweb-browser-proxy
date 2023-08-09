package helper

import "encoding/base64"

func SimpleDecoder(data []byte, encoding string) string {
	if encoding == "base64" {
		encoded := base64.StdEncoding.EncodeToString(data)
		return encoded
	}

	return string(data)
}

func SimpleEncoder(data string, encoding string) ([]byte, error) {
	if encoding == "base64" {
		return base64.StdEncoding.DecodeString(data)
	}

	return []byte(data), nil
}
