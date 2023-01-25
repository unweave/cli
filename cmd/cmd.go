package cmd

import (
	"github.com/unweave/cli/client"
	"github.com/unweave/cli/config"
)

func InitUnweaveClient() *client.Client {
	// Get token. Priority: CLI flag > Project Token > User Token
	// TODO: Implement ProjectToken parsing

	token := config.Config.Unweave.User.Token
	if config.AuthToken != "" {
		token = config.AuthToken
	}

	return client.NewClient(
		client.Config{
			ApiURL: config.Config.Unweave.ApiURL,
			Token:  token,
		})
}
