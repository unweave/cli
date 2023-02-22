package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/cobra"
	"github.com/unweave/cli/config"
	"github.com/unweave/cli/ui"
	"github.com/unweave/unweave/api/types"
)

func Login(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	uwc := InitUnweaveClient()
	code, err := uwc.Account.PairingTokenCreate(cmd.Context())
	if err != nil {
		var e *types.Error
		if errors.As(err, &e) {
			uie := &ui.Error{Error: e}
			fmt.Println(uie.Verbose())
			os.Exit(1)
			return nil
		}
		return err
	}

	authURL := config.Config.Unweave.AppURL + "/auth/pair?code=" + code
	openBrowser := ui.Confirm("Do you want to open the browser to login", "y")

	ui.Attentionf("Auth Code: %s", code)
	var openErr error
	if openBrowser {
		openErr = open.Run(authURL)
	}

	if !openBrowser || openErr != nil {
		fmt.Println("Open the following URL in your browser to login: ", authURL)
	}

	var token string
	var account *types.Account

	sleep := time.Duration(2)
	timeout := 5 * time.Minute
	retryUntil := time.Now().Add(timeout)

	for {
		if time.Now().After(retryUntil) {
			fmt.Printf("Login timed out after %f minutes \n", timeout.Minutes())
			os.Exit(1)
			return nil
		}

		token, account, err = uwc.Account.PairingTokenExchange(cmd.Context(), code)
		if err != nil {
			var e *types.Error
			if errors.As(err, &e) {
				if e.Code == http.StatusUnauthorized {
					time.Sleep(sleep * time.Second)
					continue
				}
				uie := &ui.Error{Error: e}
				fmt.Println(uie.Verbose())
				os.Exit(1)
				return nil
			}
			return err
		}
		break
	}

	config.Config.Unweave.User.Token = token
	config.Config.Unweave.User.ID = account.UserID
	config.Config.Unweave.User.Email = account.Email
	if err = config.Config.Unweave.Save(); err != nil {
		return err
	}

	ui.Successf("Logged in as %q", account.Email)
	return nil
}
