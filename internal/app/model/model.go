package model

type Product struct {
	ID              string `json:"id"`
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

func (u *User) IsSeller() bool {
	return u.Role == ROLE_SELLER
}

func (u *User) IsBuyer() bool {
	return u.Role == ROLE_BUYER
}

type TypeRole = string

const (
	ROLE_ADMIN  TypeRole = "ADMIN"
	ROLE_BUYER  TypeRole = "BUYER"
	ROLE_SELLER TypeRole = "SELLER"
)
