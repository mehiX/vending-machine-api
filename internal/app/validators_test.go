package app

import (
	"testing"
)

func TestUsernameValidate(t *testing.T) {

	type scenario struct {
		input   string
		isValid bool
	}

	scenarios := []scenario{
		{"skjdfs", false},
		{"12345678", true},
		{"ksjsksjdsk!", false},
		{"verylongusername", true},
		{"mix@valid.name-", true},
	}

	for _, s := range scenarios {
		err := validateUsername(s.input)
		if s.isValid && err != nil {
			t.Errorf("username should be valid: %s", s.input)
		}

		if !s.isValid && err == nil {
			t.Errorf("username should not pass validation: %s", s.input)
		}
	}
}

func TestPasswordValidate(t *testing.T) {
	type scenario struct {
		input   string
		isValid bool
	}

	scenarios := []scenario{
		{"12short", false},
		{"kasdfjadfjdasfkjakdsf", false},
		{"1233484744", false},
		{"AAAHAHAHAHAA", false},
		{"A678", false},
		{"hajas^&^8*&", false},
		{"123HDhhasdKJJJM", false},
		{"mhGP*&UksdfLK", false},
		{"mhG2P*&UksdfLK", true},
		{"kjadf SKS k& k7(*  ", true},
		{"    ksjE#2j    ", false},
	}

	for _, s := range scenarios {
		err := validatePassword(s.input)
		if s.isValid && err != nil {
			t.Errorf("password should be valid: %s", s.input)
		}

		if !s.isValid && err == nil {
			t.Errorf("password should not pass validation: %s", s.input)
		}
	}
}

func TestDepositValidate(t *testing.T) {
	good := []int64{0, 5, 10, 15, 20, 25, 30, 35, 50, 100, 105, 150, 2005}
	bad := []int64{1, 2, 3, 4, 6, 7, 18, 23, 22, 2501}

	for _, d := range good {
		if err := validateDeposit(d); err != nil {
			t.Errorf("unexpected error for deposit: %d. Error: %s", d, err)
		}
	}

	for _, d := range bad {
		if err := validateDeposit(d); err == nil {
			t.Errorf("deposit should be rejected: %d", d)
		}
	}
}
