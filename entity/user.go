package entity

type User struct {
	Email string `json:"email"`
}

type GeneratePairingCode struct {
	Code string `json:"code"`
}

type ExchangePairingCode struct {
	Token string `json:"token"`
}
