package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/spf13/cobra"
	"github.com/unweave/cli/config"
	"github.com/unweave/cli/ui"
	"github.com/unweave/unweave/api/types"
	"github.com/unweave/unweave/tools"
)

func dashIfZeroValue(v interface{}) interface{} {
	if v == reflect.Zero(reflect.TypeOf(v)).Interface() {
		return "-"
	}
	return v
}

// iterateSessionCreateNodeTypes attempts to create a session using the node types provided
// until the first successful creation. If none of the node types are successful, it
// returns 503 out of capacity error.
func iterateSessionCreateNodeTypes(ctx context.Context, params types.ExecCreateParams, nodeTypeIDs []string) (string, error) {
	uwc := InitUnweaveClient()

	var err error
	var session *types.Exec

	for _, nodeTypeID := range nodeTypeIDs {
		params.NodeTypeID = nodeTypeID

		owner, projectName := config.GetProjectOwnerAndName()
		session, err = uwc.Session.Create(ctx, owner, projectName, params)
		if err == nil {
			results := []ui.ResultEntry{
				{Key: "ID", Value: session.ID},
				{Key: "Provider", Value: session.Provider.DisplayName()},
				{Key: "Type", Value: session.NodeTypeID},
				{Key: "Region", Value: session.Region},
				{Key: "Status", Value: fmt.Sprintf("%s", session.Status)},
				{Key: "SSHKey", Value: fmt.Sprintf("%s", session.SSHKey.Name)},
			}

			ui.ResultTitle("Session Created:")
			ui.Result(results, ui.IndentWidth)
			return session.ID, nil
		}

		if err != nil {
			var e *types.Error
			if errors.As(err, &e) {
				// If error 503, it's mostly likely an out of capacity error. Continue to
				// next node type.
				if e.Code == 503 {
					continue
				}
				return "", err
			}
		}
	}
	// Return the last error - which will be a 503 if it's an out of capacity error.
	return "", err
}

func setupSSHKey(ctx context.Context) (string, []byte, error) {
	if config.SSHKeyName != "" {
		return config.SSHKeyName, nil, nil
	}

	user := strings.Split(config.Config.Unweave.User.Email, "@")[0]
	rsaPubGenName := fmt.Sprintf("uw:gen_%s_id_rsa", user)

	if config.SSHPublicKeyPath != "" {
		// read public key from file
		pub, err := os.ReadFile(config.SSHPublicKeyPath)
		if err != nil {
			return "", nil, err
		}
		name := filepath.Base(config.SSHPublicKeyPath)

		if name == "id_rsa.pub" {
			name = fmt.Sprintf("%s_id_rsa", user)
		}
		return name, pub, nil
	}

	if config.SSHPrivateKeyPath != "" {
		name, pub, err := sshKeyGenerateFromRSA(ctx, rsaPubGenName, config.SSHPrivateKeyPath)
		if err != nil {
			return "", nil, err
		}
		return name, pub, nil
	}

	// No key details provided, prompt user to generate new key
	ui.Attentionf("No SSH key path provided. Using default key path ~/.ssh/id_rsa")
	// Find id_rsa in ~/.ssh
	home, err := os.UserHomeDir()
	if err != nil {
		return "", nil, err
	}
	path := filepath.Join(home, ".ssh", "id_rsa")
	name, pub, err := sshKeyGenerateFromRSA(ctx, rsaPubGenName, path)
	if err != nil {
		return "", nil, err
	}
	return name, pub, nil
}

func sessionCreate(ctx context.Context, execCtx types.ExecCtx) (string, error) {
	var region *string
	var nodeTypeIDs []string

	if config.Config.Project.DefaultProvider == "" && config.Provider == "" {
		ui.Errorf("No provider specified. Either set a default provider in you project config or specify a provider with the --provider flag")
		os.Exit(1)
	}

	provider := config.Config.Project.DefaultProvider
	if config.Provider != "" {
		provider = config.Provider
	}

	if p, ok := config.Config.Project.Providers[provider]; ok {
		nodeTypeIDs = p.NodeTypes
	}
	if len(config.NodeTypeID) != 0 {
		nodeTypeIDs = []string{config.NodeTypeID}
	}
	if config.NodeRegion != "" {
		region = &config.NodeRegion
	}
	if len(nodeTypeIDs) == 0 {
		ui.Errorf("No node types specified")
		return "", fmt.Errorf("no node types specified")
	}
	if config.BuildID != "" {
		execCtx.BuildID = &config.BuildID
	}

	name, pub, err := setupSSHKey(ctx)
	if err != nil {
		return "", err
	}

	sshKeyName := &name
	sshPublicKey := tools.Stringy(string(pub))

	params := types.ExecCreateParams{
		Provider:      types.Provider(provider),
		NodeTypeID:    "",
		Region:        region,
		SSHKeyName:    sshKeyName,
		SSHPublicKey:  sshPublicKey,
		IsInteractive: true,
		Ctx:           execCtx,
	}

	sessionID, err := iterateSessionCreateNodeTypes(ctx, params, nodeTypeIDs)
	if err != nil {
		var e *types.Error
		if errors.As(err, &e) {
			if e.Code == 503 {
				// It's mostly likely an out of capacity error. Try to marshal the response
				// into a list of available node types.
				var nodeTypes []types.NodeType
				if err = json.Unmarshal([]byte(e.Suggestion), &nodeTypes); err == nil {
					cols, rows := nodeTypesToTable(nodeTypes)
					uie := &ui.Error{Error: e}
					fmt.Println(uie.Short())
					fmt.Println()
					ui.Table("Available Instances", cols, rows)
					os.Exit(1)
				}
			}
			uie := &ui.Error{Error: e}
			fmt.Println(uie.Verbose())
			return "", e
		}
		return "", err
	}

	return sessionID, nil
}

func SessionCreateCmd(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	if _, err := sessionCreate(cmd.Context(), types.ExecCtx{}); err != nil {
		os.Exit(1)
		return nil
	}
	return nil
}

func SessionList(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	uwc := InitUnweaveClient()
	listTerminated := config.All

	owner, projectName := config.GetProjectOwnerAndName()
	sessions, err := uwc.Session.List(cmd.Context(), owner, projectName, listTerminated)
	if err != nil {
		var e *types.Error
		if errors.As(err, &e) {
			uie := &ui.Error{Error: e}
			fmt.Println(uie.Verbose())
			os.Exit(1)
		}
		return err
	}

	cols := []ui.Column{
		{Title: "ID", Width: -1},
		{Title: "Provider", Width: -1},
		{Title: "Status", Width: 15},
		{Title: "Connection String", Width: -1},
	}
	rows := make([]ui.Row, len(sessions))

	for idx, s := range sessions {
		conn := "-"
		if s.Connection != nil && s.Connection.Host != "" {
			conn = fmt.Sprintf("%s@%s", s.Connection.User, s.Connection.Host)
		}
		row := ui.Row{
			fmt.Sprintf("%s", s.ID),
			fmt.Sprintf("%s", s.Provider),
			fmt.Sprintf("%s", s.Status),
			conn,
		}
		rows[idx] = row
	}

	ui.Table("Sessions", cols, rows)

	return nil
}

func sessionTerminate(ctx context.Context, sessionID string) error {
	uwc := InitUnweaveClient()
	owner, projectName := config.GetProjectOwnerAndName()
	err := uwc.Session.Terminate(ctx, owner, projectName, sessionID)
	if err != nil {
		var e *types.Error
		if errors.As(err, &e) {
			uie := &ui.Error{Error: e}
			fmt.Println(uie.Verbose())
			os.Exit(1)
		}
		return err
	}
	return nil
}

func SessionTerminate(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true
	sessionID := args[0]

	confirm := ui.Confirm(fmt.Sprintf("Are you sure you want to terminate session %q", sessionID), "n")
	if !confirm {
		return nil
	}

	if err := sessionTerminate(cmd.Context(), sessionID); err != nil {
		ui.Errorf("Failed to terminate session: %s", err.Error())
		os.Exit(1)
	}

	ui.Successf("Session terminated")
	return nil
}
