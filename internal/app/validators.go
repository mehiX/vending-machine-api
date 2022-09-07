package app

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

func validateUsername(u string) error {
	p := regexp.MustCompile(`[0-9a-zA-Z@._-]{8,}`)
	rest := p.ReplaceAllLiteralString(u, "")

	if rest != "" {
		return errors.New("username should be at least 8 characters long and may only container letters, numbers and one or more symbols @ . _ -")
	}

	return nil
}

func validatePassword(p string) error {

	trimmed := strings.TrimSpace(p)

	if len(trimmed) < password_min_length {
		return fmt.Errorf("minimum password length is %d", password_min_length)
	}

	noSmallLetters := regexp.MustCompile(`[a-z]+`).ReplaceAllLiteralString(trimmed, "")
	if len(noSmallLetters) == len(trimmed) {
		return errors.New("password should contain at least a small letter")
	}

	noCaps := regexp.MustCompile(`[A-Z]+`).ReplaceAllLiteralString(noSmallLetters, "")
	if len(noCaps) == len(noSmallLetters) {
		return errors.New("password should contain at least a capital letter")
	}

	noNumbers := regexp.MustCompile(`[0-9]+`).ReplaceAllLiteralString(noCaps, "")
	if len(noNumbers) == len(noCaps) {
		return errors.New("password should contain at least a number")
	}

	if len(noNumbers) == 0 {
		return errors.New("password should contain at least a non-alphanumerical character")
	}

	return nil
}

func validateDeposit(d int64) error {
	if d%5 != 0 {
		return errors.New("deposit should be multiple of 5")
	}

	return nil
}

func validateRole(r string) error {
	return errors.New("not implemented")
}
