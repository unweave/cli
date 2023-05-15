package config

import "github.com/unweave/cli/client"

func InitUnweaveClient() *client.Client {
	// Get token. Priority: CLI flag > Project Token > User Token
	// TODO: Implement ProjectToken parsing

	token := Config.Unweave.User.Token
	if AuthToken != "" {
		token = AuthToken
	}

	return client.NewClient(
		client.Config{
			ApiURL: Config.Unweave.ApiURL,
			Token:  token,
		})
}
