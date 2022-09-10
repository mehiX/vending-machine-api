package app

import (
	"errors"
	"fmt"
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

func TestRoleValidate(t *testing.T) {
	good := []string{"ADMIN", "BUYER", "SELLER"}
	bad := []string{"admin", "root", "", "123"}

	for _, r := range good {
		if err := validateRole(r); err != nil {
			t.Errorf("unexpected validation error for role: %s", r)
		}
	}

	for _, r := range bad {
		if err := validateRole(r); err == nil {
			t.Errorf("role should be rejected: %s", r)
		}
	}
}

func TestCostValidate(t *testing.T) {
	good := []int64{5, 10, 15, 20, 25, 30, 35, 50, 100, 105, 150, 2005}
	bad := []int64{0, 1, 2, 3, 4, 6, 7, 18, 23, 22, 2501}

	for _, c := range good {
		if err := validateCost(c); err != nil {
			t.Errorf("unexpected error for cost: %d. Error: %s", c, err)
		}
	}

	for _, c := range bad {
		if err := validateCost(c); err == nil {
			t.Errorf("cost should be rejected: %d", c)
		}
	}
}

func TestValidateDepositCoin(t *testing.T) {

	errNotAccepted := errors.New("accepted values: [5, 10. 20, 50, 100]")

	type scenario struct {
		input  int
		result error
	}

	scenarios := []scenario{
		{0, errNotAccepted},
		{-2, errNotAccepted},
		{3, errNotAccepted},
		{102, errNotAccepted},
		{25, errNotAccepted},
		{5, nil},
		{10, nil},
		{20, nil},
		{50, nil},
		{100, nil},
	}

	for _, s := range scenarios {
		t.Run(fmt.Sprintf("%d", s.input), func(t *testing.T) {
			t.Parallel()
			res := validateDepositCoin(s.input)
			if s.result == nil && res != nil {
				t.Errorf("coin should be valid: %d", s.input)
			} else if s.result != nil && res == nil {
				t.Errorf("coin should not be valid: %d", s.input)
			} else if s.result != nil && res != nil && s.result.Error() != res.Error() {
				t.Errorf("wrong error message. expected: %s, got: %s", s.result.Error(), res.Error())
			}
		})
	}
}
