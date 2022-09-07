package model

import "errors"

var (
	ErrNotImplemented  = errors.New("not implemented")
	ErrUsernameExists  = errors.New("user already exists")
	ErrInvalidPassword = errors.New("password is invalid")
)

type Product struct {
	AmountAvailable int64  `json:"amount_available"`
	Cost            int64  `json:"cost"`
	Name            string `json:"name"`
	SellerID        string `json:"seller_id"`
}

type User struct {
	ID       string
	Username string
	Password string `json:"-"`
	Deposit  int64
	Role     TypeRole
}

type TypeRole = string

const (
	ROLE_ADMIN  TypeRole = "admin"
	ROLE_USER   TypeRole = "user"
	ROLE_SELLER TypeRole = "seller"
)
