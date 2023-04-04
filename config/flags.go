package config

// All can be used across multiple commands. Example: unweave ls --all to list all projects
var All = false

// AuthToken is used to authenticate with the Unweave API. It is loaded from the saved
// config file and can be overridden with runtime flags.
var AuthToken = ""

// BuildID is the ID of the build to use when running commands that require a build.
var BuildID = ""

// NodeTypeID is the ID of the provider specific node type to use when creating a new session
var NodeTypeID = ""

// NodeRegion is the region to use when creating a new session
var NodeRegion = ""

// ProjectURI is the project slug with syntax `<owner>/<project` of the project to run
// commands on. It is loaded from the saved config file and can be overridden with runtime flags.
var ProjectURI = ""

// Provider is the provider to use when executing a request
var Provider = ""

// SSHPrivateKeyPath is the path to the SSH public key to use to connect to a new or existing Session.
var SSHPrivateKeyPath = ""

// SSHPublicKeyPath is the path to the SSH public key to use to connect to a new or existing Session.
var SSHPublicKeyPath = ""

// SSHKeyName is the name of the SSH Key already configured in Unweave to use for a new or existing Session.
var SSHKeyName = ""

// CreateSession is used to denote whether to create a new session when running commands that require a session.
var CreateSession = true
