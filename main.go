package main

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/docker/go-plugins-helpers/volume"
)

const (
	memfsId       = "_mem"
	socketAddress = "/run/docker/plugins/mem.sock"
)

var (
	defaultDir = filepath.Join(volume.DefaultDockerRootDirectory, memfsId)
	root       = flag.String("root", defaultDir, "memfs volumes root directory")
	debug      = flag.Bool("debug", false, "Enable debugging output")
)

func main() {
	flag.Parse()

	d := newMemDriver(*root, *debug)
	h := volume.NewHandler(d)

	fmt.Printf("Listening on %s\n", socketAddress)
	fmt.Println(h.ServeUnix("root", socketAddress))
}
