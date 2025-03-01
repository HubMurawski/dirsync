//go:build linux

package main

import (
	"dirsync/adapters/localfs"
	"dirsync/app/logger"
	appCfg "dirsync/config"
	"flag"
	"os"
	"runtime"
)

func main() {
	l := logger.New()
	if runtime.GOOS != "linux" {
		l.Error("%s detected! This tool supports only linux", runtime.GOOS)
		os.Exit(1)
	}

	config, err := appCfg.ParseArgs()
	if err != nil {
		l.Error("failed to parse arguments: %s", err)
		flag.Usage()
		os.Exit(1)
	}
	fsSync := localfs.NewSynchronizer(
		config.Source,
		config.Destination,
		config.DeleteMissing,
		l,
	)

	if err := fsSync.Run(); err != nil {
		l.Error("failed while running synchronizer: %s", err)
		flag.Usage()
		os.Exit(1)
	}
}
