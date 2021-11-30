package controller

// Run runs the user's latest changes and environment with Unweave. It uploads the users
// code to the server and runs it. Any files/patterns in the .gitignore file will are
// from the upload.
func (c *Controller) Run() error {
	// get the root path of the user's currently active project (user must be inside subdirectory)
	// walk filesystem at root and zip every file that's not in .gitignore
	// create a new run-session - by making a call to api.unweave.io/compute/run-session
	// upload the zip file to the api.unweave.io/compute/run-session/upload/<rid> endpoint
	return nil
}
