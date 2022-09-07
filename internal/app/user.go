package app

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/mehiX/vending-machine-api/internal/app/model"
	"golang.org/x/crypto/bcrypt"
)

const password_min_length = 8

func (a *app) CreateUser(ctx context.Context, username, password string, deposit int64, role model.TypeRole) (err error) {

	if err = validateUsername(username); err != nil {
		return
	}

	if err = validatePassword(password); err != nil {
		return
	}

	encPasswd, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return err
	}

	if a.Db == nil {
		return errors.New("no database configured")
	}

	tx, err := a.Db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			tx.Rollback()
		}
	}()

	_, err = tx.ExecContext(ctx, "insert into users (id, username, password, deposit, role) values (?, ?, ?, ?, ?)", uuid.New().String(), username, string(encPasswd), deposit, role)
	if err != nil {
		return err
	}

	return
}

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
