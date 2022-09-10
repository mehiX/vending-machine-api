package app

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/mehiX/vending-machine-api/internal/app/model"
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
	if r != model.ROLE_ADMIN && r != model.ROLE_SELLER && r != model.ROLE_BUYER {
		return fmt.Errorf("unrecognized role: %s", r)
	}

	return nil
}

func validateCost(c int64) error {
	if c == 0 || c%5 != 0 {
		return errors.New("cost is 0 or not a multiple of 5")
	}

	return nil
}

var acceptedCoinValues = []int{5, 10, 20, 50, 100}

func validateDepositCoin(v int) error {
	for _, a := range acceptedCoinValues {
		if a == v {
			return nil
		}
	}

	return fmt.Errorf("accepted values: %v", acceptedCoinValues)
}
