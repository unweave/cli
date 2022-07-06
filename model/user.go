package model

type User struct {
	Email string `json:"email"`
}

const GetMeQuery = `
	query GetMe {
		me {
			id
			email
		}
	}
`

type GeneratePairingCode struct {
	Code string `json:"code"`
}

const GeneratePairingCodeQuery = `
	mutation GeneratePairingCode{
		generatePairingCode {
			code
		}
	}
`

type ExchangePairingCode struct {
	Token string `json:"token"`
}

const ExchangePairingCodeQuery = `
	mutation ExchangePairingCode($code: String!) {
		exchangePairingCode(code: $code) {
			token
		}
	}
`
