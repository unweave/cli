package entity

import "fmt"

type UserToken struct {
	DisplayName string `json:"displayName"`
	ExpiresAt   int    `json:"expiresAt"`
	Token       string `json:"token"`
}

func (u UserToken) String() string {
	return fmt.Sprintf("DisplayName: %s, ExpiresAt: %d", u.DisplayName, u.ExpiresAt)
}

const GetUserTokensQuery = `
	query GetUserTokens {
		userTokens {
			displayName
			expiresAt
		}
	}
`

const CreateUserTokenMutation = `
	mutation CreateUserToken {
		createUserToken {
			displayName
			expiresAt
			token
		}
	}
`
