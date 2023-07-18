package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/unweave/cli/config"
	"github.com/unweave/cli/session"
	"github.com/unweave/cli/ssh"
	"github.com/unweave/cli/ui"
	"github.com/unweave/unweave/api/types"
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

	if config.Config.Project.DefaultProvider == "" && config.Provider == "" {
		ui.Errorf("No provider specified. Either set a default provider in you project config or specify a provider with the --provider flag")
		os.Exit(1)
	}
	provider := config.Config.Project.DefaultProvider
	if config.Provider != "" {
		provider = config.Provider
	}

	spec, err := parseHardwareSpec()
	if err != nil {
		return "", err
	}

	if config.NodeRegion != "" {
		region = &config.NodeRegion
	}

	if config.BuildID != "" {
		image = &config.BuildID
	}

	name, pub, err := setupSSHKey(ctx)
	if err != nil {
		return "", err
	}
	volumes, err := config.GetVolumeAttachParams()
	if err != nil {
		return "", err
	}

	params := types.ExecCreateParams{
		Provider:     types.Provider(provider),
		Spec:         spec,
		SSHKeyName:   name,
		SSHPublicKey: string(pub),
		Region:       region,
		Image:        image,
		Command:      execConfig.Command,
		CommitID:     gitConfig.CommitID,
		GitURL:       gitConfig.GitURL,
		Source:       execConfig.Src,
		Volumes:      volumes,
		InternalPort: config.InternalPort,
	}

	sessionID, err := session.Create(ctx, params)
	if err != nil {
		var e *types.Error
		if errors.As(err, &e) {
			if e.Code == 503 {
				// It's mostly likely an out of capacity error. Try to marshal the response
				// into a list of available node types.
				var nodeTypes []types.NodeType
				if err = json.Unmarshal([]byte(e.Suggestion), &nodeTypes); err == nil {
					cols, rows := gpuTypesToTable(nodeTypes)
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

	renderSessionListWithSessions(sessions)

	return nil
}

func renderSessionListWithSessions(sessions []types.Exec) {
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].Name < sessions[j].Name
	})

	if config.OutputJSON {
		if sessions == nil {
			sessions = []types.Exec{}
		}
		ui.JSON(sessions)
		return
	}

	if len(sessions) == 0 {
		ui.Infof("No active sessions")
	}

	// EITHER min length of title + 5 for padding OR the max field length + 5 for padding
	cols := []ui.Column{
		{
			Title: "Name",
			Width: 3 + ui.MaxFieldLength(sessions, func(exec types.Exec) string {
				return exec.Name
			}),
		},
		{
			Title: "Provider",
			Width: 5 + ui.MaxFieldLength(sessions, func(exec types.Exec) string {
				return exec.Provider.String()
			}),
		},
		{
			Title: "Instance Type",
			Width: 7 + ui.MaxFieldLength(sessions, func(exec types.Exec) string {
				return string(exec.Spec.GPU.Type)
			}),
		},
		{
			Title: "Status",
			Width: 5 + ui.MaxFieldLength(sessions, func(exec types.Exec) string {
				return string(exec.Status)
			}),
		},
	}

	rows := make([]ui.Row, len(sessions))

	for idx, s := range sessions {
		instanceType := s.Spec.GPU.Type
		if instanceType == "" {
			instanceType = s.Spec.CPU.Type
		}

		row := ui.Row{
			fmt.Sprintf("%s", s.Name),
			fmt.Sprintf("%s", s.Provider),
			fmt.Sprintf("%s", instanceType),
			fmt.Sprintf("%s", s.Status),
		}
		rows[idx] = row
	}

	ui.Table("Sessions", cols, rows)
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
func sessionSelectSSHExecRef(ctx context.Context, execRef string, allowNew bool) (execID string, isNewSession bool, err error) {
	const newSessionOpt = "✨  Create a new session"

	execs, err := getExecs(ctx)
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

	execRef, err = renderCobraSelection(ctx, cobraOpts, selectionIdByIdx, "Select a session to connect to")
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
		txt := fmt.Sprintf("%s - %s - %s - (%s)", s.Name, s.Provider, s.ID, s.Status)
		options[len(prepend)+idx] = txt
		optionMap[len(prepend)+idx] = s.ID
	}

	return options, optionMap
}

func parseHardwareSpec() (types.HardwareSpec, error) {
	specName := "default"
	if config.SpecName != "" {
		specName = config.SpecName
	}

	spec, found := findSpec(specName, config.Config.Project.Specs)
	if !found {
		return types.HardwareSpec{}, &types.Error{
			Message:    fmt.Sprintf("cannot find spec with name %q", specName),
			Suggestion: fmt.Sprintf("ensure a spec with name %q exists in .unweave/config.toml", specName),
		}
	}

	baseHardwaveSpec := types.HardwareSpec{
		GPU: types.GPU{
			Type:  spec.GPU.Type,
			Count: types.HardwareRequestRange{Min: spec.GPU.Count, Max: spec.GPU.Count},
			RAM:   types.HardwareRequestRange{Min: spec.GPU.Memory, Max: spec.GPU.Memory},
		},
		CPU: types.CPU{
			Type:                 spec.CPU.Type,
			HardwareRequestRange: types.HardwareRequestRange{Min: spec.CPU.Count, Max: spec.CPU.Count},
		},
		RAM: types.HardwareRequestRange{Min: spec.CPU.Memory, Max: spec.CPU.Memory},
		HDD: types.HardwareRequestRange{Min: spec.HDD.Size, Max: spec.HDD.Size},
	}

	setNotEmptyValue(config.GPUType, &(baseHardwaveSpec.GPU.Type))
	setNotEmptyValue(config.GPUs, &(baseHardwaveSpec.GPU.Count.Max))
	setNotEmptyValue(config.GPUs, &(baseHardwaveSpec.GPU.Count.Min))
	setNotEmptyValue(config.GPUMemory, &(baseHardwaveSpec.GPU.RAM.Max))
	setNotEmptyValue(config.GPUMemory, &(baseHardwaveSpec.GPU.RAM.Min))
	setNotEmptyValue(config.CPUs, &(baseHardwaveSpec.CPU.HardwareRequestRange.Max))
	setNotEmptyValue(config.CPUs, &(baseHardwaveSpec.CPU.HardwareRequestRange.Min))
	setNotEmptyValue(config.Memory, &(baseHardwaveSpec.RAM.Max))
	setNotEmptyValue(config.Memory, &(baseHardwaveSpec.RAM.Min))
	setNotEmptyValue(config.HDD, &(baseHardwaveSpec.HDD.Max))
	setNotEmptyValue(config.HDD, &(baseHardwaveSpec.HDD.Min))

	return baseHardwaveSpec, nil
}

func findSpec(name string, specs []config.Spec) (config.Spec, bool) {
	for _, s := range specs {
		if strings.EqualFold(s.Name, name) {
			return s, true
		}
	}

	return config.Spec{}, false
}

func setNotEmptyValue[T comparable](val T, dst *T) {
	if val == *new(T) {
		return
	}

	*dst = val
}
