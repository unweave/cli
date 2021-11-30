package entity

type User struct {
	Id    int64  `json:"id"`
	Email string `json:"email"`
}

type GeneratePairingCode struct {
	Code string `json:"code"`
}

type ExchangePairingCode struct {
	Uid   string `json:"userId"`
	Token string `json:"token"`
}
