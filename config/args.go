package cfg

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

func ParseArgs() (Args, error) {
	flag.Usage = func() {
		fmt.Println("usage: dirsync [--delete-missing] [-src] <source> [-dst] <destination>")
		flag.PrintDefaults()
	}

	src := flag.String("src", "", "Source directory for synchronization process")
	dst := flag.String("dst", "", "Destination directory for synchronization process")
	deleteMissing := flag.Bool("delete-missing", false, "Deletes all files from destination directory that are not present in source directory")
	flag.Parse()

	if src == nil || *src == "" {
		*src = flag.Arg(0)
	}
	if src == nil || *src == "" {
		return Args{}, errors.New("source cannot be empty")
	}
	if dst == nil || *dst == "" {
		*dst = flag.Arg(1)
	}
	if dst == nil || *dst == "" {
		return Args{}, errors.New("destination cannot be empty")
	}
	if _, err := os.Stat(*src); os.IsNotExist(err) {
		return Args{}, errors.New("source directory does not exist: " + *src)
	}

	return Args{
		Source:        *src,
		Destination:   *dst,
		DeleteMissing: *deleteMissing,
	}, nil
}

type Args struct {
	Source        string
	Destination   string
	DeleteMissing bool
}
