package main

import (
	"flag"
	"fmt"

	"github.com/forewing/gobuild"
)

const (
	name = "csgo-rcon"
)

var (
	flagAll = flag.Bool("all", false, "build for all platforms")

	target = gobuild.Target{
		Source:      "./cmd/csgo-rcon",
		OutputName:  name,
		OutputPath:  "./output",
		CleanOutput: true,

		ExtraFlags:   []string{"-trimpath"},
		ExtraLdFlags: "-s -w",

		VersionPath: "",
		HashPath:    "",

		Compress:  gobuild.CompressRaw,
		Platforms: []gobuild.Platform{{}},
	}
)

func main() {
	flag.Parse()
	if *flagAll {
		target.OutputName = fmt.Sprintf("%s-%s-%s-%s",
			name,
			gobuild.PlaceholderVersion,
			gobuild.PlaceholderOS,
			gobuild.PlaceholderArch)
		target.Compress = gobuild.CompressZip
		target.Platforms = gobuild.PlatformCommon
	}
	err := target.Build()
	if err != nil {
		panic(err)
	}
}
