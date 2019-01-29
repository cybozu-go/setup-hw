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

func outputJSON(hash, salt []byte) error {
	v := struct {
		Hash string `json:"hash"`
		Salt string `json:"salt"`
	}{
		Hash: toHex(hash),
		Salt: toHex(salt),
	}

	return json.NewEncoder(os.Stdout).Encode(v)
}

func outputPlain(hash, salt []byte) error {
	_, err := os.Stdout.WriteString(
		fmt.Sprintf("Hash: %s\nSalt: %s\n", toHex(hash), toHex(salt)))
	return err
}
