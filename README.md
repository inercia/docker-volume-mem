# Docker memfs volume plugin

[![Build Status](https://drone.io/github.com/inercia/docker-volume-mem/status.png)](https://drone.io/github.com/inercia/docker-volume-mem/latest)

Mounts a in-memory filesystem inside your contaniner.

The FUSE mount point is shared between containers if the name of the volume is the
same between containers. Otherwhise, a new volume is mounted per container.

## Installation

You could download a pre-built binary from the [drone.io](http://drone.io)
build [here](https://drone.io/github.com/inercia/docker-volume-mem/files/docker-volume-mem),
but the recommended procedure would be by using `go`:

```
$ go get github.com/inercia/docker-volume-mem
```

## Usage

1. Run the daemon:

```
$ sudo docker-volume-mem
```

2. Run containers pointing to the driver:

```
$ docker run --volume-driver mem -v mymem:/shared --name c1 -it alpine /bin/ash
```

The `/shared` folder will be in container `c1` and shared with any other container
using the `mymem` volume id. So starting a second container with:

```
$ docker run --volume-driver mem -v mymem:/shared --name c2 -it alpine /bin/ash
```

will results in both containers sharing an in-memory folder mounted at `/shared`.

## Systemd socket activation

The plugin can be socket activated by _systemd_. You must `make install` for
installing the files provided under `systemd/`. This ensures the plugin gets activated
if for some reasons it's down.

## LICENSE

MIT
