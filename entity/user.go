package entity

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
	mutation {
		generatePairingCode {
			code
		}
	}
`

type ExchangePairingCode struct {
	Token string `json:"token"`
}

const ExchangePairingCodeQuery = `
	query($token: String!) {
		exchangePairingCode(token: $token) {
			token
		}
	}
`
