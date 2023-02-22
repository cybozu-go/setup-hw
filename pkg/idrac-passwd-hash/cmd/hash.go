package cmd

import (
	"crypto/rand"
	"crypto/sha256"
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

func generateSalt() ([]byte, error) {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}
	return salt, nil
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
