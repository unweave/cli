package entity

import (
	"fmt"
	"time"
)

type UserToken struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	ExpiresAt   int64  `json:"expiresAt"`
	Token       string `json:"token"`
}

func (u UserToken) String() string {
	expires := time.UnixMilli(u.ExpiresAt)
	expiresStr := expires.Format("2006-01-02 15:04:05")
	return fmt.Sprintf("ID: %s, DisplayName: %s, ExpiresAt: %s", u.ID, u.DisplayName, expiresStr)
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
