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
			t.Errorf("username '%s' should be valid", s.input)
		}

		if !s.isValid && err == nil {
			t.Errorf("username '%s' should not pass validation", s.input)
		}
	}
}
