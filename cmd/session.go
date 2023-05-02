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
	"time"

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
	if config.SSHPublicKeyPath != "" {
		path := strings.Replace(config.SSHPublicKeyPath, ".pub", "", 1)
		return sshKeyAddIDRSA(ctx, path, nil)
	}

	if config.SSHPrivateKeyPath != "" {
		name, pub, err := sshKeyGenerateFromPrivateKey(ctx, config.SSHPrivateKeyPath, nil)
		if err != nil {
			return "", nil, err
		}
		return name, pub, nil
	}

	// No key details provided, try using ~/.ssh/id_rsa.pub
	ui.Attentionf("No SSH key path provided. Using default key path ~/.ssh/id_rsa")

	home, err := os.UserHomeDir()
	if err != nil {
		return "", nil, err
	}
	path := filepath.Join(home, ".ssh", "id_rsa")

	idrsaExists := false
	idrsaPubExists := false

	if _, err := os.Stat(path); err == nil {
		idrsaPubExists = true
	}
	if _, err := os.Stat(path + ".pub"); err == nil {
		idrsaExists = true
	}

	if idrsaExists && idrsaPubExists {
		name, pub, err := sshKeyAddIDRSA(ctx, path, nil)
		if err != nil {
			return "", nil, err
		}
		return name, pub, nil
	}
	ui.Attentionf("No SSH key found at %s", path)
	ui.Attentionf("Generating new SSH key")

	genName, pub, err := sshKeyGenerate(ctx, config.Config.Unweave.User.ID, nil)
	if err != nil {
		return "", nil, err
	}

	return genName, pub, nil
}

func sessionCreate(ctx context.Context, execConfig types.ExecConfig, gitConfig types.GitConfig) (string, error) {
	var region, image *string
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
		image = &config.BuildID
	}

	name, pub, err := setupSSHKey(ctx)
	if err != nil {
		return "", err
	}

	sshKeyName := &name
	sshPublicKey := tools.Stringy(string(pub))

	params := types.ExecCreateParams{
		Provider:     types.Provider(provider),
		NodeTypeID:   "",
		Region:       region,
		SSHKeyName:   sshKeyName,
		SSHPublicKey: sshPublicKey,
		Image:        image,
		Command:      execConfig.Command,
		CommitID:     gitConfig.CommitID,
		GitURL:       gitConfig.GitURL,
		Source:       execConfig.Src,
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

func sessionCreateAndWatch(ctx context.Context, execConfig types.ExecConfig, gitConfig types.GitConfig) (exech chan types.Exec, errch chan error, err error) {
	sessionID, err := sessionCreate(ctx, execConfig, gitConfig)
	if err != nil {
		ui.Errorf("Failed to create session: %v", err)
		os.Exit(1)
		return nil, nil, nil
	}

	uwc := InitUnweaveClient()
	listTerminated := config.All
	owner, projectName := config.GetProjectOwnerAndName()

	errch = make(chan error)
	exech = make(chan types.Exec)
	currentStatus := types.StatusInitializing

	go func() {
		ticketCount := 0
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				sessions, err := uwc.Session.List(ctx, owner, projectName, listTerminated)
				if err != nil {
					var e *types.Error
					if errors.As(err, &e) {
						uie := &ui.Error{Error: e}
						fmt.Println(uie.Verbose())
						os.Exit(1)
					}
					errch <- err
					return
				}

				for _, s := range sessions {
					s := s
					if s.ID == sessionID {
						if s.Status != currentStatus {
							currentStatus = s.Status
							exech <- s
						}
						if s.Status == types.StatusError {
							ui.Errorf("❌ Session %s failed to start", sessionID)
							os.Exit(1)
						}
						if s.Status == types.StatusTerminated {
							ui.Errorf("Session %q is terminated.", sessionID)
							os.Exit(1)
						}

						if ticketCount%10 == 0 && s.Status != types.StatusRunning {
							ui.Infof("Waiting for session %q to start...", sessionID)
						}
						ticketCount++
					}
				}

			case <-ctx.Done():
				return
			}
		}
	}()

	return exech, errch, nil
}

func SessionCreateCmd(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	if _, err := sessionCreate(cmd.Context(), types.ExecConfig{}, types.GitConfig{}); err != nil {
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

func sessionTerminate(ctx context.Context, execID string) error {
	uwc := InitUnweaveClient()
	owner, projectName := config.GetProjectOwnerAndName()

	err := uwc.Session.Terminate(ctx, owner, projectName, execID)
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

	var execID string

	if len(args) == 1 {
		execID = args[0]
	}

	if len(args) == 0 {
		uwc := InitUnweaveClient()
		listTerminated := config.All

		owner, projectName := config.GetProjectOwnerAndName()
		execs, err := uwc.Session.List(cmd.Context(), owner, projectName, listTerminated)
		if err != nil {
			var e *types.Error
			if errors.As(err, &e) {
				uie := &ui.Error{Error: e}
				fmt.Println(uie.Verbose())
				os.Exit(1)
			}
			return err
		}

		optionMap := make(map[int]string)
		options := make([]string, len(execs))

		for idx, s := range execs {
			txt := fmt.Sprintf("%s - %s - %s - (%s)", s.Name, s.Provider, s.NodeTypeID, s.Status)
			options[idx] = txt
			optionMap[idx] = s.ID
		}

		selected, err := ui.Select("Select session to terminate", options)
		if err != nil {
			return err
		}

		execID = optionMap[selected]
	}

	confirm := ui.Confirm(fmt.Sprintf("Are you sure you want to terminate session %q", execID), "n")
	if !confirm {
		return nil
	}

	if err := sessionTerminate(cmd.Context(), execID); err != nil {
		ui.Errorf("Failed to terminate session: %s", err.Error())
		os.Exit(1)
	}

	ui.Successf("Session terminated")
	return nil
}
