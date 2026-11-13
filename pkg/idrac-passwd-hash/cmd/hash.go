package cmd

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"fmt"
	"os"

	"golang.org/x/term"
)

func readPasswordFromStdTerminal(prompt string) (string, error) {
	if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
		return "", fmt.Errorf("stdin and stdout are not terminals")
	}

	fmt.Print(prompt)
	p, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return "", err
	}
	return string(p), nil
}

func askPassword() ([]byte, error) {
	pass, err := readPasswordFromStdTerminal("Enter password: ")
	if err != nil {
		return nil, err
	}
	pass2, err := readPasswordFromStdTerminal("Retype password: ")
	if err != nil {
		return nil, err
	}
	if pass != pass2 {
		return nil, errors.New("password mismatch")
	}

	return []byte(pass), nil
}

const (
	// iDRAC9 and earlier
	saltSize = 16

	// iDRAC10
	idrac10SaltSize   = 32
	idrac10Iterations = 1000
	idrac10KeySize    = 32
)

func generateSalt(size int) ([]byte, error) {
	salt := make([]byte, size)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}
	return salt, nil
}

// hashPasswordIDRAC10 returns the value of iDRAC.Users.N.PasswordHashSalt for iDRAC10:
//
//	base64(PBKDF2-HMAC-SHA384(password, salt, 1000 iterations, 32 bytes)) + "|" + base64(salt)
func hashPasswordIDRAC10(passwd, salt []byte) (string, error) {
	key, err := pbkdf2.Key(sha512.New384, string(passwd), salt, idrac10Iterations, idrac10KeySize)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key) + "|" + base64.StdEncoding.EncodeToString(salt), nil
}

func hashPassword(passwd, salt []byte) ([]byte, error) {
	h := sha256.New()
	_, err := h.Write(passwd)
	if err != nil {
		return nil, err
	}
	_, err = h.Write(salt)
	if err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}
