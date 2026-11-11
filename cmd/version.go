package cmd

import "runtime/debug"

// version is set at build time with:
//
//	-ldflags "-X github.com/CosmosDevops/servicemeow/cmd.version=v1.2.3"
var version = "dev"

func init() {
	// `go install github.com/CosmosDevops/servicemeow@v1.2.3` doesn't set
	// ldflags, but records the module version in the build info.
	if version == "dev" {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
			version = info.Main.Version
		}
	}
	rootCmd.Version = version
}
