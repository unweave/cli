package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/google/uuid"
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
func iterateSessionCreateNodeTypes(ctx context.Context, nodeTypeIDs []string, region, sshKeyName, sshPublicKey *string) (uuid.UUID, error) {
	uwc := InitUnweaveClient()

	var err error
	var session *types.Session

	for _, nodeTypeID := range nodeTypeIDs {
		params := types.SessionCreateParams{
			Provider:     types.RuntimeProvider(config.Config.Project.DefaultProvider),
			NodeTypeID:   nodeTypeID,
			Region:       region,
			SSHKeyName:   sshKeyName,
			SSHPublicKey: sshPublicKey,
		}

		projectID := config.Config.Project.ID
		session, err = uwc.Session.Create(ctx, projectID, params)
		if err == nil {
			results := []ui.ResultEntry{
				{Key: "ID", Value: session.ID.String()},
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
			var e *types.HTTPError
			if errors.As(err, &e) {
				// If error 503, it's mostly likely an out of capacity error. Continue to
				// next node type.
				if e.Code == 503 {
					continue
				}
				uie := &ui.Error{HTTPError: e}
				fmt.Println(uie.Verbose())
				return uuid.Nil, e
			}
		}
	}
	// Return the last error - which will be a 503 if it's an out of capacity error.
	return uuid.Nil, err
}

func sessionCreate(ctx context.Context) (uuid.UUID, error) {
	var region *string
	var nodeTypeIDs []string

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
		return uuid.Nil, fmt.Errorf("no node types specified")
	}

	sshKeyName := tools.Stringy("")
	sshPublicKey := tools.Stringy("")

	// Use key name in request is provided, otherwise try reading public key from file
	if config.SSHKeyName != "" {
		sshKeyName = &config.SSHKeyName
	} else {
		if config.SSHKeyPath == "" {
			newKey := ui.Confirm("No SSH key path provided. Do you want to generate a new SSH key", "n")
			if !newKey {
				ui.Errorf("No SSH key path provided")
				return uuid.Nil, fmt.Errorf("no ssh key path provided")
			}

			name, path, err := sshKeyGenerate(ctx, nil)
			if err != nil {
				return uuid.Nil, err
			}
			sshKeyName = &name
			config.SSHKeyPath = path
		}

		f, err := os.ReadFile(config.SSHKeyPath)
		if err != nil {
			ui.Errorf("Failed to read public key file: %s", err.Error())
			os.Exit(1)
		}
		s := string(f)
		sshPublicKey = &s
		sshKeyName = tools.Stringy(filepath.Base(config.SSHKeyPath))
	}

	sessionID, err := iterateSessionCreateNodeTypes(ctx, nodeTypeIDs, region, sshKeyName, sshPublicKey)
	if err != nil {
		var e *types.HTTPError
		if errors.As(err, &e) {
			if e.Code == 503 {
				// It's mostly likely an out of capacity error. Try to marshal the response
				// into a list of available node types.
				var nodeTypes []types.NodeType
				if err = json.Unmarshal([]byte(e.Suggestion), &nodeTypes); err == nil {
					cols, rows := nodeTypesToTable(nodeTypes)
					uie := &ui.Error{HTTPError: e}
					fmt.Println(uie.Short())
					fmt.Println()
					ui.Table("Available Instances", cols, rows)
					os.Exit(1)
				}
			}
			uie := &ui.Error{HTTPError: e}
			fmt.Println(uie.Verbose())
			return uuid.Nil, e
		}
		return uuid.Nil, err
	}

	return sessionID, nil
}

func SessionCreateCmd(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true
	if _, err := sessionCreate(cmd.Context()); err != nil {
		os.Exit(1)
		return nil
	}
	return nil
}

func SessionList(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	uwc := InitUnweaveClient()
	listTerminated := config.All

	projectID := config.Config.Project.ID
	sessions, err := uwc.Session.List(cmd.Context(), projectID, listTerminated)
	if err != nil {
		var e *types.HTTPError
		if errors.As(err, &e) {
			uie := &ui.Error{HTTPError: e}
			fmt.Println(uie.Verbose())
			os.Exit(1)
		}
		return err
	}

	cols := []ui.Column{{Title: "ID", Width: 38}, {Title: "Status", Width: 15}}
	rows := make([]ui.Row, len(sessions))

	for idx, s := range sessions {
		row := ui.Row{fmt.Sprintf("%s", s.ID), fmt.Sprintf("%s", s.Status)}
		rows[idx] = row
	}

	ui.Table("Sessions", cols, rows)

	return nil
}

func SessionTerminate(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	sessionID, err := uuid.Parse(args[0])
	if err != nil {
		fmt.Println("Invalid session ID")
		return nil
	}

	confirm := ui.Confirm(fmt.Sprintf("Are you sure you want to terminate session %q", sessionID), "n")
	if !confirm {
		return nil
	}

	uwc := InitUnweaveClient()
	projectID := config.Config.Project.ID
	err = uwc.Session.Terminate(cmd.Context(), projectID, sessionID)
	if err != nil {
		var e *types.HTTPError
		if errors.As(err, &e) {
			uie := &ui.Error{HTTPError: e}
			fmt.Println(uie.Verbose())
			os.Exit(1)
		}
		return err
	}

	ui.Successf("Session terminated")
	return nil
}
