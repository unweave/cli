package config

// All can be used across multiple commands. Example: unweave ls --all to list all projects
var All = false

// AuthToken is used to authenticate with the Unweave API. It is loaded from the saved
// config file and can be overridden with runtime flags.
var AuthToken = ""

// BuildID is the ID of the build to use when running commands that require a build.
var BuildID = ""

// CreateExec is used to denote whether to create a new exec when running commands that require a exec.
var CreateExec = true

// GPUs is the number of GPUs to allocate for a gpuType.
var GPUs int

// GPUMemory is the memory of GPU if applicable for a gpuType.
var GPUMemory int

// GPUType is the type of GPU to use.
var GPUType string

// CPUs is the number of VCPUs to allocate.
var CPUs int

// Memory is the amount of RAM to allocate in GB.
var Memory int

// HDD is the amount of storage to allocate in GB.
var HDD int

// NodeRegion is the region to use when creating a new session
var NodeRegion = ""

// InternalPort is the port that should be exposed as https
var InternalPort int32

// ProjectURI is the project slug with syntax `<owner>/<project` of the project to run
// commands on. It is loaded from the saved config file and can be overridden with runtime flags.
var ProjectURI = ""

// Provider is the provider to use when executing a request
var Provider = ""

// SSHPrivateKeyPath is the path to the SSH public key to use to connect to a new or existing Exec.
var SSHPrivateKeyPath = ""

// SSHPublicKeyPath is the path to the SSH public key to use to connect to a new or existing Exec.
var SSHPublicKeyPath = ""

// SSHKeyName is the name of the SSH Key already configured in Unweave to use for a new or existing Exec.
var SSHKeyName = ""

// SSHConnectionOptions is the arguments you want to include when opening an SSH session.
var SSHConnectionOptions []string

// NoCopySource is a bool to denote whether to copy the source code to the session
var NoCopySource = true

// Volumes is a list of volumes to mount to the session
var Volumes []string
