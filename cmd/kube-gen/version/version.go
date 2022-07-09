package version

import (
	"runtime/debug"
)

var (
	// Version is the current version
	Version = ""

	// GitCommit is the commit hash from which the binary was built
	GitCommit = ""

	// BuildTime is the time at which the binary was built
	BuildTime = ""
)

func init() {
	if Version != "" && GitCommit != "" {
		// build info is set via ldflags
		return
	}

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}

	var dirty string
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			GitCommit = s.Value
			if len(GitCommit) >= 9 {
				GitCommit = GitCommit[:9]
			}
		case "vcs.modified":
			if s.Value == "true" {
				dirty = "-dirty"
			}
		case "vcs.time":
			if BuildTime == "" {
				BuildTime = s.Value
			}
		}
	}

	if Version == "" {
		Version = "unknown"
	}

	GitCommit += dirty
}
