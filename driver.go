package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/docker/go-plugins-helpers/volume"
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"golang.org/x/sys/unix"
)

type mount struct {
	name     string
	server   *fuse.Server
	refCount int
}

type memDriver struct {
	sync.Mutex

	root   string
	mounts map[string]*mount
	debug  bool
}

func newMemDriver(root string, debug bool) memDriver {
	// Locks memory, preventing memory from being written to disk as swap
	err := unix.Mlockall(unix.MCL_FUTURE | unix.MCL_CURRENT)
	switch err {
	case nil:
	case unix.ENOSYS:
		log.Println("WARNING: mlockall() not implemented on this system")
	case unix.ENOMEM:
		log.Println("WARNING: mlockall() failed with ENOMEM")
	default:
		log.Fatalf("FATAL: could not perform mlockall and prevent swapping memory: %v", err)
	}

	return memDriver{
		root:   root,
		mounts: map[string]*mount{},
		debug:  debug,
	}
}

func (d memDriver) Create(r volume.Request) volume.Response {
	d.log("Creating %s", r.Name)
	return volume.Response{}
}

func (d memDriver) Get(r volume.Request) volume.Response {
	d.Lock()
	defer d.Unlock()

	m := d.mountpoint(r.Name)
	if s, ok := d.mounts[m]; ok {
		return volume.Response{Volume: &volume.Volume{Name: s.name, Mountpoint: d.mountpoint(s.name)}}
	}

	return volume.Response{Err: fmt.Sprintf("Unable to find volume mounted on %s", m)}
}

func (d memDriver) List(r volume.Request) volume.Response {
	d.Lock()
	defer d.Unlock()

	var vols []*volume.Volume
	for _, v := range d.mounts {
		vols = append(vols, &volume.Volume{Name: v.name, Mountpoint: d.mountpoint(v.name)})
	}
	return volume.Response{Volumes: vols}
}

func (d memDriver) Remove(r volume.Request) volume.Response {
	d.Lock()
	defer d.Unlock()

	d.log("Removing %s", r.Name)
	mountpoint := d.mountpoint(r.Name)
	if s, ok := d.mounts[mountpoint]; ok {
		if s.refCount <= 1 {
			delete(d.mounts, mountpoint)
		}
	}

	return volume.Response{}
}

func (d memDriver) Path(r volume.Request) volume.Response {
	return volume.Response{Mountpoint: d.mountpoint(r.Name)}
}

func (d memDriver) Mount(r volume.Request) volume.Response {
	d.Lock()
	defer d.Unlock()

	d.log("Mounting %s", r.Name)
	mountpoint := d.mountpoint(r.Name)
	log.Printf("Mounting volume %s on %s\n", r.Name, mountpoint)

	m, ok := d.mounts[mountpoint]
	if ok && m.refCount > 0 {
		log.Printf("Volume %s already mounted on %s\n", r.Name, mountpoint)
		m.refCount++
		return volume.Response{Mountpoint: mountpoint}
	}

	fi, err := os.Lstat(mountpoint)

	if os.IsNotExist(err) {
		if err := os.MkdirAll(mountpoint, 0755); err != nil {
			return volume.Response{Err: err.Error()}
		}
	} else if err != nil {
		return volume.Response{Err: err.Error()}
	}
	if fi != nil && !fi.IsDir() {
		return volume.Response{Err: fmt.Sprintf("%v already exist and it's not a directory", mountpoint)}
	}
	if err := os.MkdirAll(filepath.Dir(mountpoint), 0755); err != nil {
		return volume.Response{Err: err.Error()}
	}

	mountOptions := &fuse.MountOptions{
		AllowOther: true,
		Name:       "mem",
		Options:    []string{"default_permissions"},
	}

	root := nodefs.NewMemNodeFSRoot(r.Name)
	conn := nodefs.NewFileSystemConnector(root, &nodefs.Options{})
	server, err := fuse.NewServer(conn.RawFS(), mountpoint, mountOptions)
	if err != nil {
		fmt.Printf("Mount fail: %v\n", err)
		os.Exit(1)
	}
	server.SetDebug(d.debug)
	fmt.Println("Mounted!")

	d.mounts[mountpoint] = &mount{
		name:     mountpoint,
		server:   server,
		refCount: 1,
	}

	go server.Serve()

	return volume.Response{Mountpoint: mountpoint}
}

func (d memDriver) Unmount(r volume.Request) volume.Response {
	d.Lock()
	defer d.Unlock()

	mountpoint := d.mountpoint(r.Name)
	log.Printf("Unmounting volume %s from %s\n", r.Name, mountpoint)

	if m, ok := d.mounts[mountpoint]; ok {
		if m.refCount == 1 {
			m.server.Unmount()
		}
		m.refCount--
	} else {
		return volume.Response{Err: fmt.Sprintf("Unable to find volume mounted on %s", mountpoint)}
	}

	return volume.Response{}
}

func (d *memDriver) mountpoint(name string) string {
	return filepath.Join(d.root, name)
}

func (d *memDriver) log(format string, args ...interface{}) {
	if (d.debug) {
		log.Printf(format, args...)
	}
}
