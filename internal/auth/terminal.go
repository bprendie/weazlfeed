package auth

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/bprendie/weazlfeed/internal/store"
	"golang.org/x/term"
)

func UnlockOrCreate(vault *store.Store) error {
	has, err := vault.HasLock()
	if err != nil {
		return err
	}
	if has {
		password, err := readSecret("Vault password: ")
		if err != nil {
			return err
		}
		return vault.Unlock(password)
	}
	password, err := readSecret("Create vault password: ")
	if err != nil {
		return err
	}
	if strings.TrimSpace(password) == "" {
		return errors.New("password is required")
	}
	confirm, err := readSecret("Confirm vault password: ")
	if err != nil {
		return err
	}
	if password != confirm {
		return errors.New("passwords do not match")
	}
	return vault.CreateLock(password)
}

func readSecret(prompt string) (string, error) {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return "", errors.New("password unlock requires an interactive terminal")
	}
	fmt.Fprint(os.Stderr, prompt)
	value, err := term.ReadPassword(fd)
	fmt.Fprintln(os.Stderr)
	return string(value), err
}
