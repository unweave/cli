package controller

import (
	"context"
	"fmt"
)

func (c *Controller) GetUser(ctx context.Context, id int64, email string) error {
	vars := struct {
		Id    int64  `json:"id"`
		Email string `json:"email"`
	}{
		Id:    id,
		Email: email,
	}
	req, err := c.api.NewGqlRequest(`
		query GetUser ($id: BigInt, $email: String) {
			user (email: $email, id: $id) {
				id
				email
			}
		}`, vars)

	if err != nil {
		return err
	}

	var resp struct {
		User struct {
			Id    *int64  `json:"id"`
			Email *string `json:"email"`
		} `json:"user"`
	}

	err = c.api.ExecuteGql(ctx, req, &resp)
	if err != nil {
		return err
	}

	fmt.Println("resp", *resp.User.Id, *resp.User.Email)
	return nil
}

func (c *Controller) LoginWithToken(ctx context.Context, token string) error {
	return nil
}

func (c *Controller) LoginWithBrowser(ctx context.Context) error {
	code, err := c.api.GeneratePairingCode(ctx)
	if err != nil {
		return err
	}

	fmt.Println("code", code)
	return nil
}
