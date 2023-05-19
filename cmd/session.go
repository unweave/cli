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
	"github.com/unweave/cli/session"
	"github.com/unweave/cli/ssh"
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

func setupSSHKey(ctx context.Context) (string, []byte, error) {
	if config.SSHKeyName != "" {
		return config.SSHKeyName, nil, nil
	}

	if config.SSHPublicKeyPath != "" {
		path := strings.Replace(config.SSHPublicKeyPath, ".pub", "", 1)
		return sshKeyAddIDRSA(ctx, path, nil)
	}

	if config.SSHPrivateKeyPath != "" {
		name, pub, err := ssh.GenerateFromPrivateKey(ctx, config.SSHPrivateKeyPath, nil)
		if err != nil {
			return "", nil, err
		}
		return name, pub, nil
	}

	// No key details provided, try using ~/.unweave_global/.ssh/
	name, pub, err := getFirstPublicKeyInPath(ctx, config.GetUnweaveSSHKeysFolder())
	if err == nil {
		return name, pub, nil
	}

	// No SSH key found, generate a new one
	return generateSSHKey(ctx)
}

// Deprecated: we no longer use the idrsa key as default
func useIdRsaSSHKey(ctx context.Context) (string, []byte, error) {
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
		ui.Infof("Using default key path ~/.ssh/id_rsa")
		return name, pub, nil
	}

	return "", nil, fmt.Errorf("No SSH key found at %s", path)
}

func generateSSHKey(ctx context.Context) (string, []byte, error) {
	dir := config.GetUnweaveSSHKeysFolder()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".pub") {
			continue
		}

		config.SSHPublicKeyPath = filepath.Join(dir, entry.Name())

		keyname, err := ssh.Add(ctx, config.SSHPublicKeyPath, config.Config.Unweave.User.ID, nil)
		if err != nil {
			return "", nil, err
		}
		ui.Infof("Using default key path %s", config.SSHPublicKeyPath)
		return keyname, nil, nil
	}

	ui.Attentionf("No SSH key found at %s", dir)
	ui.Attentionf("Generating new SSH key")

	name, keypath, pub, err := ssh.Generate(ctx, config.Config.Unweave.User.ID, nil)
	if err != nil {
		return "", nil, err
	}

	ui.Successf("Created new SSH key pair:\n"+
		"  Name: %s\n"+
		"  Path: %s\n",
		name, keypath)

	return name, pub, nil
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

	sessionID, err := session.Create(ctx, params, nodeTypeIDs)
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

func execCreateAndWatch(ctx context.Context, execConfig types.ExecConfig, gitConfig types.GitConfig) (exech chan types.Exec, errch chan error, err error) {
	execID, err := sessionCreate(ctx, execConfig, gitConfig)
	if err != nil {
		ui.Errorf("Failed to create session: %v", err)
		os.Exit(1)
		return nil, nil, nil
	}
	return session.Wait(ctx, execID)
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

	uwc := config.InitUnweaveClient()
	listTerminated := config.All

	owner, projectName := config.GetProjectOwnerAndName()
	sessions, err := uwc.Exec.List(cmd.Context(), owner, projectName, listTerminated)
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
	uwc := config.InitUnweaveClient()
	owner, projectName := config.GetProjectOwnerAndName()

	err := uwc.Exec.Terminate(ctx, owner, projectName, execID)
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
		execs, err := getExecs(cmd.Context())
		if err != nil {
			var e *types.Error
			if errors.As(err, &e) {
				uie := &ui.Error{Error: e}
				fmt.Println(uie.Verbose())
				os.Exit(1)
			}
			return err
		}

		opts, execIdByOpts := formatExecCobraOpts(execs)
		execID, _ = renderCobraSelection(cmd.Context(), opts, execIdByOpts, "Select session to terminate")

		if len(execs) == 0 {
			ui.Attentionf("No active sessions found")
			return nil
		}
	}

	if execID == "" {
		// This shouldn't really happen
		ui.Attentionf("No session selected")
		return nil
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

// sessionSelectSSHExecRef selects an exec id from all sessions in the Unweave environment or whether to create a new
// provides an option to create a new Exec an error or exits if unrecoverable
func sessionSelectSSHExecRef(cmd *cobra.Command, execRef string, allowNew bool) (execID string, isNewSession bool, err error) {
	const newSessionOpt = "✨  Create a new session"

	execs, err := getExecs(cmd.Context())
	if err != nil {
		if e, ok := err.(*types.Error); ok {
			uie := &ui.Error{Error: e}
			fmt.Println(uie.Verbose())
			os.Exit(1)
		}
		return "", false, err
	}

	var cobraOpts = make([]string, 0, len(execs))
	var selectionIdByIdx = make(map[int]string, len(execs))

	if allowNew {
		cobraOpts, selectionIdByIdx = formatExecCobraOpts(execs)
	} else {
		cobraOpts, selectionIdByIdx = formatExecCobraOpts(execs, newSessionOpt)
	}

	execRef, err = renderCobraSelection(cmd.Context(), cobraOpts, selectionIdByIdx, "Select a session to connect to")
	if err != nil {
		return "", false, err
	}

	return execRef, execRef == newSessionOpt, nil
}

// formatExecCobraOpts returns Cobra options per exec, and a map associating option idx to its Exec ID
// prepends any additional options in prepend
func formatExecCobraOpts(execs []types.Exec, prepend ...string) ([]string, map[int]string) {
	optionMap := make(map[int]string)
	options := make([]string, len(prepend)+len(execs))

	for idx, opt := range prepend {
		options[idx] = opt
		optionMap[idx] = opt
	}

	for idx, s := range execs {
		txt := fmt.Sprintf("%s - %s - %s - (%s)", s.Name, s.Provider, s.NodeTypeID, s.Status)
		options[len(prepend)+idx] = txt
		optionMap[len(prepend)+idx] = s.ID
	}

	return options, optionMap
}
