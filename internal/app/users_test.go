package app

import "testing"

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
