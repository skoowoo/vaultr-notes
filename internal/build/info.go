package build

import "fmt"

// These variables are injected at build time via -ldflags.
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

// Info holds build metadata.
type Info struct {
	Version   string
	Commit    string
	BuildDate string
}

// Get returns the current build info.
func Get() Info {
	return Info{
		Version:   Version,
		Commit:    Commit,
		BuildDate: BuildDate,
	}
}

func (i Info) String() string {
	return fmt.Sprintf("version=%s commit=%s buildDate=%s", i.Version, i.Commit, i.BuildDate)
}
