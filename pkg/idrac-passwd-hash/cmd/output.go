package cmd

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func toHex(data []byte) string {
	return strings.ToUpper(hex.EncodeToString(data))
}

func outputJSON(hash, salt []byte, hashSalt string) error {
	v := struct {
		Hash     string `json:"hash"`
		Salt     string `json:"salt"`
		HashSalt string `json:"hash_salt"`
	}{
		Hash:     toHex(hash),
		Salt:     toHex(salt),
		HashSalt: hashSalt,
	}

	return json.NewEncoder(os.Stdout).Encode(v)
}

func outputPlain(hash, salt []byte, hashSalt string) error {
	_, err := os.Stdout.WriteString(
		fmt.Sprintf("Hash: %s\nSalt: %s\nHashSalt: %s\n", toHex(hash), toHex(salt), hashSalt))
	return err
}
