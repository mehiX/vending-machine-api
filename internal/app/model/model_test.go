package model

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestUser(t *testing.T) {

	seller := &User{Role: ROLE_SELLER}
	buyer := &User{Role: ROLE_BUYER}
	admin := &User{Role: ROLE_ADMIN}

	if !seller.IsSeller() || buyer.IsSeller() || admin.IsSeller() {
		t.Error("IsSeller doesn't work")
	}

	if !buyer.IsBuyer() || seller.IsBuyer() || admin.IsBuyer() {
		t.Error("IsBuyer doesn't work")
	}
}

func TestUserAsJsonDoesNotExposePassword(t *testing.T) {

	usr := &User{
		Password: "some password",
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(usr); err != nil {
		t.Fatal(err)
	}

	str := strings.ToLower(buf.String())
	if strings.Contains(str, "pass") {
		t.Fatal("encoded string contains the password")
	}
}
