package entity

import (
	"encoding/json"
	"fmt"
	"time"
)

type UserToken struct {
	ID          string    `json:"id"`
	DisplayName string    `json:"displayName"`
	ExpiresAt   time.Time `json:"expiresAt"`
	Token       string    `json:"token"`
}

func (u *UserToken) UnmarshalJSON(b []byte) error {
	type raw struct {
		ID          string          `json:"id"`
		DisplayName string          `json:"displayName"`
		ExpiresAt   json.RawMessage `json:"expiresAt"`
		Token       string          `json:"token"`
	}
	var r raw
	if err := json.Unmarshal(b, &r); err != nil {
		return err
	}

	var timeStamp int64
	if err := json.Unmarshal(r.ExpiresAt, &timeStamp); err != nil {
		return err
	}

	u.ID = r.ID
	u.DisplayName = r.DisplayName
	u.ExpiresAt = time.UnixMilli(timeStamp)
	u.Token = r.Token

	return nil
}

func (u *UserToken) String() string {
	expiresStr := u.ExpiresAt.Format("2006-01-02 15:04:05")
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
