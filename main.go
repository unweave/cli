package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/muesli/reflow/wordwrap"
	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/cobra"
	"github.com/unweave/cli/cmd"
	"github.com/unweave/cli/config"
	"github.com/unweave/cli/ui"
	"github.com/unweave/cli/vars"
	"github.com/unweave/unweave/api/types"
)

var (
	groupDev        = "dev"
	groupManagement = "management"
	repoOwner       = "unweave"
	repoName        = "cli"

	rootCmd = &cobra.Command{
		Use:           "unweave <command>",
		Short:         "Create serverless sessions to train your ML models",
		Args:          cobra.MinimumNArgs(0),
		SilenceUsage:  false,
		SilenceErrors: false,
	}
)

type RunE func(cmd *cobra.Command, args []string) error
type Release struct {
	TagName string `json:"tag_name"`
}

func withValidProjectURI(r RunE) RunE {
	return func(cmd *cobra.Command, args []string) error {
		if config.ProjectURI == "" && config.Config.Project.URI == "" {
			ui.Errorf("No project ID set. Either run `unweave link` first or use the `--project` flag to set a project ID.")
			os.Exit(1)
		}
		if config.ProjectURI != "" {
			// Override project ID if set via flag
			config.Config.Project.URI = config.ProjectURI
		}
		return r(cmd, args)
	}
}

func verifyCLIVersion(currentVersion, latestVersion string) {
	// Don't check for updates if we're running a dev version
	if latestVersion == "dev" || currentVersion == "dev" {
		return
	}

	if latestVersion != currentVersion {
		ui.Attentionf("Your unweave CLI is out of date. Yours: %s, Latest: %s.", currentVersion, latestVersion)
		ui.Attentionf("To update, run: brew update && brew upgrade unweave")
	}
}

func getLatestReleaseVersion(owner, repo string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	var release Release
	err = json.Unmarshal(body, &release)
	if err != nil {
		return "", err
	}

	v := strings.TrimPrefix(release.TagName, "v")
	return v, nil
}

func init() {
	rootCmd.Version = config.Version
	rootCmd.Flags().BoolP("version", "v", false, "Get the version of current Unweave CLI")
	rootCmd.AddGroup(&cobra.Group{ID: groupDev, Title: "Dev:"})
	rootCmd.AddGroup(&cobra.Group{ID: groupManagement, Title: "Account Management:"})

	flags := rootCmd.PersistentFlags()
	flags.StringVar(&config.ProjectURI, "project", "", "Use a specific project ID - overrides config")
	flags.StringVarP(&config.AuthToken, "token", "t", "", "Use a specific token to authenticate - overrides login token")
	flags.BoolVar(&vars.Debug, "debug", false, "Enable debug mode")

	flags.StringVarP(&config.SSHKeyName, "key", "k", "", "Name of the SSH key to use")
	flags.StringVar(&config.SSHPublicKeyPath, "pub", "", "Path to the SSH public key to use")
	flags.BoolVar(&config.OutputJSON, "json", false, "Output JSON instead of human-readable text")

	rootCmd.AddCommand(&cobra.Command{
		Use:     "build [path]",
		Short:   "Build a project into a container image",
		GroupID: groupDev,
		Args:    cobra.RangeArgs(0, 1),
		RunE:    withValidProjectURI(cmd.Build),
		Hidden:  true,
	})

	boxCmd := &cobra.Command{
		Use:     "box [box-name]",
		Short:   "Create a new session with a persistent filesystem",
		GroupID: groupDev,
		RunE:    withValidProjectURI(cmd.BoxUp),
		Args:    cobra.RangeArgs(0, 1),
		Hidden:  true,
	}
	boxCmd.Flags().StringVar(&config.Provider, "provider", "", "Provider to use")
	boxCmd.Flags().StringVar(&config.NodeRegion, "region", "", "Region to use, eg. `us_west_2`")

	rootCmd.AddCommand(boxCmd)

	codeCmd := &cobra.Command{
		Use:     "code",
		Short:   "Create a new session and open it in VS Code",
		GroupID: groupDev,
		Args:    cobra.RangeArgs(0, 1),
		RunE:    withValidProjectURI(cmd.Code),
	}
	codeCmd.Flags().BoolVar(&config.CreateExec, "new", false, "Create a new")
	codeCmd.Flags().StringVarP(&config.BuildID, "image", "i", "", "Build ID of the container image to use")
	codeCmd.Flags().StringVar(&config.Provider, "provider", "", "Provider to use")
	codeCmd.Flags().StringVar(&config.NodeRegion, "region", "", "Region to use, eg. `us_west_2`")
	codeCmd.Flags().StringVar(&config.SSHPrivateKeyPath, "prv", "", "Absolute Path to the private key to use")
	codeCmd.Flags().IntVar(&config.GPUs, "gpus", 0, "Number of GPUs to allocate for a gpuType, e.g., 2")
	codeCmd.Flags().IntVar(&config.GPUMemory, "gpu-mem", 0, "Memory of GPU if applicable for a gpuType, e.g., 12")
	codeCmd.Flags().StringVar(&config.GPUType, "gpu-type", "", "Type of GPU to use, e.g., rtx_5000")
	codeCmd.Flags().IntVar(&config.CPUs, "cpus", 0, "Number of VCPUs to allocate, e.g., 4")
	codeCmd.Flags().IntVar(&config.Memory, "mem", 0, "Amount of RAM to allocate in GB, e.g., 16")
	codeCmd.Flags().IntVar(&config.HDD, "hdd", 0, "Amount of hard-disk space to allocate in GB")
	codeCmd.Flags().StringVar(&config.SpecName, "spec", "default", "Spec from config to use")
	codeCmd.Flags().StringSliceVarP(&config.Volumes, "volume", "v", []string{}, "Mount a volume to the exec. e.g., -v <volume-name>:/data")
	codeCmd.Flags().Int32VarP(&config.InternalPort, "port", "p", 0, "Port on the exec to expose as an https interface e.g. -p 8080")

	rootCmd.AddCommand(codeCmd)

	cpCmd := &cobra.Command{
		Use:   "cp <source-path> <destination-path>",
		Short: "Copy files and folders to or from a remote host",
		Long: wordwrap.String("Copy files and folders to or from a remote host \n\n"+
			"To copy to a remote session, you must prefix the remote path with the session "+
			"name `sess:<session-name>`. The `sess:` prefix is necessary.\n\n"+
			"The full syntax for this command is:\n\n"+
			"Local to remote:\n"+
			"unweave cp <local-path> sess:<session-name><remote-path>\n\n"+
			"Remote to local:\n"+
			"unweave cp sess:<session-name><remote-path> <local-path> \n\n"+
			"Current directory to remote:\n"+
			"unweave cp . sess:<session-name><remote-path>\n\n"+
			"Example: \n"+
			"\tunweave cp /home/data sess:session-name/home/ml-data\n"+
			"\tunweave cp sess:session-name/home/ml-data /home/data\n", ui.MaxOutputLineLength),
		Args:    cobra.ExactArgs(2),
		Aliases: []string{"cp", "copy"},
		GroupID: groupDev,
		RunE:    withValidProjectURI(cmd.Copy),
	}
	rootCmd.AddCommand(cpCmd)

	rootCmd.AddCommand(&cobra.Command{
		Use:     "config",
		Short:   "Show the current config",
		GroupID: groupDev,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(config.Config.String())
		},
	})
	execCmd := &cobra.Command{
		GroupID: groupDev,
		Short:   "Execute a command serverlessly",
		Use:     "exec [flags] -- [<command>]...",
		RunE:    withValidProjectURI(cmd.Exec),
	}
	execCmd.Flags().StringVarP(&config.BuildID, "image", "i", "", "Build ID of the container image to use")
	execCmd.Flags().StringVar(&config.Provider, "provider", "", "Provider to use")
	execCmd.Flags().StringVar(&config.NodeRegion, "region", "", "Region to use, eg. `us_west_2`")
	execCmd.Flags().IntVar(&config.GPUs, "gpus", 0, "Number of GPUs to allocate for a gpuType, e.g., 2")
	execCmd.Flags().IntVar(&config.GPUMemory, "gpu-mem", 0, "Memory of GPU if applicable for a gpuType, e.g., 12")
	execCmd.Flags().StringVar(&config.GPUType, "gpu-type", "", "Type of GPU to use, e.g., rtx_5000")
	execCmd.Flags().IntVar(&config.CPUs, "cpus", 0, "Number of VCPUs to allocate, e.g., 4")
	execCmd.Flags().IntVar(&config.Memory, "mem", 0, "Amount of RAM to allocate in GB, e.g., 16")
	execCmd.Flags().IntVar(&config.HDD, "hdd", 0, "Amount of hard-disk space to allocate in GB")
	execCmd.Flags().StringVar(&config.SpecName, "spec", "default", "Spec from config to use")
	execCmd.Flags().StringSliceVarP(&config.Volumes, "volume", "v", []string{}, "Mount a volume to the exec. e.g., -v <volume-name>:/data")
	execCmd.Flags().Int32VarP(&config.InternalPort, "port", "p", 0, "Port on the exec to expose as an https interface e.g. -p 8080")
	execCmd.Flags().BoolVar(&config.ExecAttach, "interactive", false, "Stay attached in an interactive terminal session to the exec after starting the command")
	execCmd.Flags().StringSliceVar(&config.SSHConnectionOptions, "connection-option", []string{}, "SSH connection config to include e.g StrictHostKeyChecking=yes")
	execCmd.Flags().BoolVar(&config.NoCopySource, "no-copy", false, "Do not copy source code to the session")

	rootCmd.AddCommand(execCmd)

	logsCmd := &cobra.Command{
		GroupID: groupDev,
		Short:   "Print logs from an exec",
		Use:     "logs [flags] [session-id|name]",
		RunE:    withValidProjectURI(cmd.Logs),
	}
	logsCmd.Flags().StringSliceVar(&config.SSHConnectionOptions, "connection-option", []string{}, "SSH connection config to include e.g StrictHostKeyChecking=yes")
	logsCmd.Flags().BoolVarP(&config.FollowLogs, "follow", "f", false, "Stream logs as they are created")

	rootCmd.AddCommand(logsCmd)

	linkCmd := &cobra.Command{
		Use:     "link [project-id]",
		Aliases: []string{"init"}, // this is temp
		Short:   "Init a new project config in your local directory and link it to an Unweave project",
		GroupID: groupManagement,
		Args:    cobra.ExactArgs(1),
		RunE:    cmd.Link,
	}
	linkCmd.Flags().StringP("path", "p", "", "Path to the project directory")
	rootCmd.AddCommand(linkCmd)

	// Auth
	loginCmd := &cobra.Command{
		Use:     "login",
		Short:   "Login to Unweave",
		GroupID: groupManagement,
		RunE:    cmd.Login,
	}
	rootCmd.AddCommand(loginCmd)

	rootCmd.AddCommand(&cobra.Command{
		Use:    "logout",
		Short:  "Logout of Unweave",
		RunE:   cmd.Logout,
		Hidden: true,
	})

	// Provider commands
	lsNodeType := &cobra.Command{
		Use:   "ls-gpu-types <provider>",
		Short: "List gpu types available on a provider",
		Args:  cobra.ExactArgs(1),
		RunE:  cmd.ProviderListGPUTypes,
	}
	lsNodeType.Flags().BoolVarP(&config.All, "all", "a", false, "Including out of capacity gpu types")
	rootCmd.AddCommand(lsNodeType)

	rootCmd.AddCommand(&cobra.Command{
		Use:   "open",
		Short: "Open the Unweave dashboard in your browser",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := open.Run(config.Config.Unweave.AppURL); err != nil {
				ui.Errorf("Failed to open browser: %s", err)
				os.Exit(1)
			}
			return nil
		},
		Args: cobra.NoArgs,
	})

	// Session commands
	newCmd := &cobra.Command{
		Use:   "new",
		Short: "Create a new Unweave session.",
		Long: wordwrap.String("Create a new Unweave session. If no region is provided,"+
			"the first available one will be selected.", ui.MaxOutputLineLength),
		Args:    cobra.NoArgs,
		GroupID: groupDev,
		RunE:    withValidProjectURI(cmd.SessionCreateCmd),
	}
	newCmd.Flags().StringVarP(&config.BuildID, "image", "i", "", "Build ID of the container image to use")
	newCmd.Flags().StringVar(&config.Provider, "provider", "", "Provider to use")
	newCmd.Flags().StringVar(&config.NodeRegion, "region", "", "Region to use, eg. `us_west_2`")
	newCmd.Flags().IntVar(&config.GPUs, "gpus", 0, "Number of GPUs to allocate for a gpuType, e.g., 2")
	newCmd.Flags().IntVar(&config.GPUMemory, "gpu-mem", 0, "Memory of GPU if applicable for a gpuType, e.g., 12")
	newCmd.Flags().StringVar(&config.GPUType, "gpu-type", "", "Type of GPU to use, e.g., rtx_5000")
	newCmd.Flags().IntVar(&config.CPUs, "cpus", 0, "Number of VCPUs to allocate, e.g., 4")
	newCmd.Flags().IntVar(&config.Memory, "mem", 0, "Amount of RAM to allocate in GB, e.g., 16")
	newCmd.Flags().IntVar(&config.HDD, "hdd", 0, "Amount of hard-disk space to allocate in GB")
	newCmd.Flags().StringVar(&config.SpecName, "spec", "default", "Spec from config to use")
	newCmd.Flags().StringSliceVarP(&config.Volumes, "volume", "v", []string{}, "Mount a volume to the exec. e.g., -v <volume-name>:/data")
	newCmd.Flags().Int32VarP(&config.InternalPort, "port", "p", 0, "Port on the exec to expose as an https interface e.g. -p 8080")

	rootCmd.AddCommand(newCmd)

	lsCmd := &cobra.Command{
		Use:     "ls",
		Short:   "List active Unweave sessions",
		Long:    "List active Unweave sessions. To list all sessions, use the --all flag.",
		Args:    cobra.NoArgs,
		Aliases: []string{"list"},
		GroupID: groupDev,
		RunE:    withValidProjectURI(cmd.SessionList),
	}
	lsCmd.Flags().BoolVarP(&config.All, "all", "a", false, "List all sessions")
	rootCmd.AddCommand(lsCmd)

	rootCmd.AddCommand(&cobra.Command{
		Use:     "terminate [session-id]",
		Short:   "Terminate an Unweave session",
		Args:    cobra.RangeArgs(0, 1),
		Aliases: []string{"delete", "del"},
		GroupID: groupDev,
		RunE:    withValidProjectURI(cmd.SessionTerminate),
	})

	sshCmd := &cobra.Command{
		Use:   "ssh [session-name|id]",
		Short: "SSH into existing session or create a new one",
		//Long: "You can specify arguments for the ssh command after a double dash (--). \n" +
		//	"For example: \n" +
		//	"	`unweave ssh -- -L 8080:localhost:8080`\n" +
		//	"   `unweave ssh <session-name|id> -- -L 8080:localhost:8080`\n",
		GroupID: groupDev,
		RunE:    withValidProjectURI(cmd.SSH),
	}
	sshCmd.Flags().BoolVar(&config.CreateExec, "new", false, "Create a new session")
	sshCmd.Flags().BoolVar(&config.NoCopySource, "no-copy", false, "Do not copy source code to the session")
	sshCmd.Flags().StringVarP(&config.BuildID, "image", "i", "", "Build ID of the container image to use")
	sshCmd.Flags().StringVar(&config.Provider, "provider", "", "Provider to use")
	sshCmd.Flags().StringVar(&config.NodeRegion, "region", "", "Region to use, eg. `us_west_2`")
	sshCmd.Flags().StringVar(&config.SSHPrivateKeyPath, "prv", "", "Absolute Path to the private key to use")
	sshCmd.Flags().IntVar(&config.GPUs, "gpus", 0, "Number of GPUs to allocate for a gpuType, e.g., 2")
	sshCmd.Flags().IntVar(&config.GPUMemory, "gpu-mem", 0, "Memory of GPU if applicable for a gpuType, e.g., 12")
	sshCmd.Flags().StringVar(&config.GPUType, "gpu-type", "", "Type of GPU to use, e.g., rtx_5000")
	sshCmd.Flags().IntVar(&config.CPUs, "cpus", 0, "Number of VCPUs to allocate, e.g., 4")
	// Setting RAM causes issues right now
	sshCmd.Flags().IntVar(&config.Memory, "mem", 0, "Amount of RAM to allocate in GB, e.g., 16")
	sshCmd.Flags().IntVar(&config.HDD, "hdd", 0, "Amount of hard-disk space to allocate in GB")
	sshCmd.Flags().StringVar(&config.SpecName, "spec", "default", "Spec from config to use")
	sshCmd.Flags().StringSliceVarP(&config.Volumes, "volume", "v", []string{}, "Mount a volume to newly created execs. e.g., -v <volume-name>:/data")
	sshCmd.Flags().Int32VarP(&config.InternalPort, "port", "p", 0, "Port on the exec to expose as an https interface e.g. -p 8080")

	rootCmd.AddCommand(sshCmd)

	// SSH Key commands
	sshKeyCmd := &cobra.Command{
		Use:     "ssh-keys",
		Aliases: []string{"ssh-key", "sshkey"},
		Short:   "Manage Unweave SSH keys: add | generate | ls",
		GroupID: groupDev,
		Args:    cobra.NoArgs,
	}
	sshKeyCmd.AddCommand(&cobra.Command{
		Use:   "add <public-key-path> [name]",
		Short: "Add a new SSH key to Unweave",
		Args:  cobra.RangeArgs(1, 2),
		RunE:  cmd.SSHKeyAdd,
	})
	sshKeyCmd.AddCommand(&cobra.Command{
		Use:   "generate [name]",
		Short: "Generate a new SSH key and add it to Unweave",
		Args:  cobra.RangeArgs(0, 1),
		RunE:  cmd.SSHKeyGenerate,
	})
	sshKeyCmd.AddCommand(&cobra.Command{
		Use:   "ls",
		Short: "List Unweave SSH keys",
		Args:  cobra.NoArgs,
		RunE:  cmd.SSHKeyList,
	})
	rootCmd.AddCommand(sshKeyCmd)

	// Volume commands
	volumeCmd := &cobra.Command{
		Use:     "volume",
		Short:   "Manage volumes in Unweave",
		GroupID: groupDev,
		Aliases: []string{"vol"},
		Args:    cobra.NoArgs,
	}

	volumeNewCmd := &cobra.Command{
		Use:   "new <name> [size (GB]]",
		Short: "Create a new volume in Unweave",
		Long: wordwrap.String("Create a new volume in Unweave.\n\n"+
			"Eg. unweave volume new my-vol 100\n\n"+
			"The volume name must be unique per project. \n",
			ui.MaxOutputLineLength),
		Args:    cobra.RangeArgs(1, 2),
		Aliases: []string{"new", "n", "create", "c"},
		RunE:    cmd.VolumeCreate,
	}
	volumeNewCmd.Flags().StringVar(&config.Provider, "provider", types.UnweaveProvider.String(), "Provider to use")
	volumeCmd.AddCommand(volumeNewCmd)

	volumeCmd.AddCommand(&cobra.Command{
		Use:   "ls",
		Short: "List volumes",
		Args:  cobra.NoArgs,
		RunE:  cmd.VolumeList,
	})

	volumeCmd.AddCommand(&cobra.Command{
		Use:   "resize <name> <size-in-gb>",
		Short: "Resize a volume",
		Args:  cobra.RangeArgs(1, 2),
		RunE:  cmd.VolumeResize,
	})

	rootCmd.AddCommand(volumeCmd)

	endpointCommand := &cobra.Command{
		Use:     "endpoint",
		Short:   "Manage endpoints in Unweave",
		GroupID: groupDev,
		Aliases: []string{"end"},
		Args:    cobra.NoArgs,
	}

	endpointNewCmd := &cobra.Command{
		Use:   "new <exec-id>",
		Short: "Create a new endpoint in Unweave",
		Long: wordwrap.String("Create a new endpoint in Unweave.\n\n"+
			"Eg. unweave endpoint new <exec-id>\n\n",
			ui.MaxOutputLineLength),
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"new", "n", "create", "c"},
		RunE:    cmd.EndpointCreate,
	}
	endpointCommand.AddCommand(endpointNewCmd)

	endpointListCmd := &cobra.Command{
		Use:   "list",
		Short: "List endpoints in Unweave",
		Long: wordwrap.String("List endpoints in Unweave.\n\n"+
			"Eg. unweave endpoint list\n\n",
			ui.MaxOutputLineLength),
		Aliases: []string{"ls"},
		RunE:    cmd.EndpointList,
	}
	endpointCommand.AddCommand(endpointListCmd)

	endpointEvalCheckCmd := &cobra.Command{
		Use:   "check <endpoint-id>",
		Short: "Check endpoints in Unweave",
		Long: wordwrap.String("Check endpoints in Unweave.\n\n"+
			"Eg. unweave endpoint check <endpoint-id>\n\n",
			ui.MaxOutputLineLength),
		Args: cobra.ExactArgs(1),
		RunE: cmd.EndpointEvalCheck,
	}
	endpointCommand.AddCommand(endpointEvalCheckCmd)

	endpointEvalCheckStatusCmd := &cobra.Command{
		Use:   "check-status <check-id>",
		Short: "Status of endpoint checks in Unweave",
		Long: wordwrap.String("Status of endpoint checks Unweave.\n\n"+
			"Eg. unweave endpoint check-status <check-id>\n\n",
			ui.MaxOutputLineLength),
		Args: cobra.ExactArgs(1),
		RunE: cmd.EndpointCheckStatus,
	}
	endpointCommand.AddCommand(endpointEvalCheckStatusCmd)

	endpointEvalAttachCmd := &cobra.Command{
		Use:   "attach-eval <endpoint-id> <eval-id>",
		Short: "Attach evals to endpoints in Unweave",
		Long: wordwrap.String("Attach evals to endpoints in Unweave.\n\n"+
			"Eg. unweave endpoint attach-eval <endpoint-id> <eval-id>\n\n",
			ui.MaxOutputLineLength),
		Args: cobra.ExactArgs(2),
		RunE: cmd.EndpointEvalAttach,
	}
	endpointCommand.AddCommand(endpointEvalAttachCmd)
	rootCmd.AddCommand(endpointCommand)

	evalCommand := &cobra.Command{
		Use:     "eval",
		Short:   "Manage evals in Unweave",
		GroupID: groupDev,
		Aliases: []string{"eval"},
		Args:    cobra.NoArgs,
	}

	evalNewCmd := &cobra.Command{
		Use:   "new <exec-id>",
		Short: "Create a new eval in Unweave",
		Long: wordwrap.String("Create a new eval in Unweave.\n\n"+
			"Eg. unweave eval new <exec-id>\n\n",
			ui.MaxOutputLineLength),
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"new", "n", "create", "c"},
		RunE:    cmd.EvalCreate,
	}
	evalCommand.AddCommand(evalNewCmd)

	evalListCmd := &cobra.Command{
		Use:   "list",
		Short: "List evals in Unweave",
		Long: wordwrap.String("List evals in Unweave.\n\n"+
			"Eg. unweave eval list\n\n",
			ui.MaxOutputLineLength),
		Aliases: []string{"ls"},
		RunE:    cmd.EvalList,
	}
	evalCommand.AddCommand(evalListCmd)
	rootCmd.AddCommand(evalCommand)

	deployCmd := &cobra.Command{
		Use:   "deploy [flags] [filepath]",
		Short: "Deploy the project",
		Long: wordwrap.String("Deploy the project in Unweave.\n\n"+
			"e.g. unweave deploy .\n\n",
			ui.MaxOutputLineLength),
		RunE: cmd.Deploy,
	}
	deployCmd.Flags().StringVarP(&config.BuildID, "image", "i", "", "Build ID of the container image to use")
	deployCmd.Flags().StringVar(&config.Provider, "provider", "", "Provider to use")
	deployCmd.Flags().StringVar(&config.NodeRegion, "region", "", "Region to use, eg. `us_west_2`")
	deployCmd.Flags().IntVar(&config.GPUs, "gpus", 0, "Number of GPUs to allocate for a gpuType, e.g., 2")
	deployCmd.Flags().IntVar(&config.GPUMemory, "gpu-mem", 0, "Memory of GPU if applicable for a gpuType, e.g., 12")
	deployCmd.Flags().StringVar(&config.GPUType, "gpu-type", "", "Type of GPU to use, e.g., rtx_5000")
	deployCmd.Flags().IntVar(&config.CPUs, "cpus", 0, "Number of VCPUs to allocate, e.g., 4")
	deployCmd.Flags().IntVar(&config.Memory, "mem", 0, "Amount of RAM to allocate in GB, e.g., 16")
	deployCmd.Flags().IntVar(&config.HDD, "hdd", 0, "Amount of hard-disk space to allocate in GB")
	deployCmd.Flags().StringVar(&config.SpecName, "spec", "default", "Spec from config to use")
	deployCmd.Flags().StringVar(&config.Command, "cmd", "", "Command to run in the deploy [required]")
	deployCmd.Flags().StringSliceVarP(&config.Volumes, "volume", "v", []string{}, "Mount a volume to the exec. e.g., -v <volume-name>:/data")
	deployCmd.Flags().Int32VarP(&config.InternalPort, "port", "p", 8080, "Port on the exec to expose as an https interface e.g. -p 8080")
	deployCmd.Flags().StringVar(&config.EndpointName, "endpoint", "", "name of the endpoint to deploy")
	deployCmd.Flags().StringSliceVar(&config.SSHConnectionOptions, "connection-option", []string{}, "SSH connection config to include e.g StrictHostKeyChecking=yes")

	rootCmd.AddCommand(deployCmd)
}

func main() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	defer func() {
		if r := recover(); r != nil {
			// TODO: Send telemetry
			ui.Errorf("Aw snap 😣 Something went wrong! %v", r)
			os.Exit(1)
		}
	}()

	config.Init()

	currentVersion := config.Version
	latestVersion, err := getLatestReleaseVersion(repoOwner, repoName)
	if err != nil {
		ui.Errorf("Failed to check latest CLI version")
	} else {
		verifyCLIVersion(currentVersion, latestVersion)
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
