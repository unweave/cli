package version

var Version = "Inserted through ldflags"

func GetVersion() string {
	return Version
}
