package app

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

type TypeRole int

const (
	ROLE_ADMIN TypeRole = iota
	ROLE_USER
	ROLE_SELLER
)
