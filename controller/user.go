package controller

import (
	"context"
	goErr "errors"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/skratchdot/open-golang/open"
	"github.com/unweave/cli/api"
	"github.com/unweave/cli/entity"
	"github.com/unweave/cli/errors"
	"time"
)

func (c *Controller) LoginWithToken(ctx context.Context, token string) error {
	return nil
}

func (c *Controller) LoginWithBrowser(ctx context.Context) error {
	code, err := c.api.GeneratePairingCode(ctx)
	if err != nil {
		fmt.Printf("Ow snap ☠️ failed to generate pairing code")
		return err
	}

	authUrl := api.GetAppUrl() + "/auth/pair?code=" + code
	prompt := &survey.Confirm{
		Message: "Do you want to open the browser to login?",
		Default: true,
	}

	shouldOpen := true
	if err = survey.AskOne(prompt, &shouldOpen); err != nil {
		return err
	}

	var openErr error
	if shouldOpen {
		openErr = open.Run(authUrl)
	}

	if !shouldOpen || openErr != nil {
		fmt.Println("Please open the following URL in your browser: ", authUrl)
	}

	var uid, token string
	sleep := time.Duration(2)
	timeout := 5 * time.Minute
	retryUntil := time.Now().Add(timeout)

	for {
		if time.Now().After(retryUntil) {
			return fmt.Errorf("login timed out after %f minutes", timeout.Minutes())
		}

		uid, token, err = c.api.ExchangePairingCode(ctx, code)
		if goErr.Is(err, errors.HttpUnAuthorized) {
			// Hasn't yet been paired
			time.Sleep(sleep * time.Second)
			continue
		}
		break
	}

	err = c.cfg.UpdateUserConfig(entity.UserConfig{
		Id:    uid,
		Token: token,
	})
	if err != nil {
		return err
	}
	return nil
}
