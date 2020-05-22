package domain

type Account struct {
	ID          string `json:"id"`
	HasPassword bool   `json:"has_password"`
	Password    string
}
