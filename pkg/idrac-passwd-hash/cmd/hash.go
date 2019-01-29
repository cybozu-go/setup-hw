package cmd

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"os"

	"github.com/howeyc/gopass"
)

func askPassword() ([]byte, error) {
	pass, err := gopass.GetPasswdPrompt("Enter password: ", false, os.Stdin, os.Stdout)
	if err != nil {
		return nil, err
	}
	pass2, err := gopass.GetPasswdPrompt("Retype password: ", false, os.Stdin, os.Stdout)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(pass, pass2) {
		return nil, errors.New("password mismatch")
	}

	return pass, nil
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
