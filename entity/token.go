package entity

import "fmt"

type UserToken struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	ExpiresAt   int    `json:"expiresAt"`
	Token       string `json:"token"`
}

func (u UserToken) String() string {
	return fmt.Sprintf("ID: %s, DisplayName: %s, ExpiresAt: %d", u.ID, u.DisplayName, u.ExpiresAt)
}

const GetUserTokensQuery = `
	query GetUserTokens {
		userTokens {
			id
			displayName
			expiresAt
		}
	}
`

const CreateUserTokenMutation = `
	mutation CreateUserToken {
		createUserToken {
			id
			displayName
			expiresAt
			token
		}
	}
`
