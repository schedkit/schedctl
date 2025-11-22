# schedctl

`schedctl` lets you run sched_ext-powered userspace schedulers packaged inside OCI images.

## Installation

### openSUSE Tumbleweed
`schedctl` is available in openSUSE Factory:

```sh
sudo zypper in schedctl
```

### Arch Linux
`schedctl` is available on AUR, and you and install it using your favorite AUR helper:

```sh
paru -S schedctl
```

## Container engine setup

### Podman

In case you want to use Podman as your container engine of choice, you need to start the Podman socket to make sure `schedctl` can connect to it.

```sh
sudo systemctl start podman.socket
```

### containerd

In case you want to use containerd as your container engine of choice, you just need to start the service.

```sh
sudo systemctl start containerd
```

## Usage

Starting and stopping a scheduler using schedctl is trivial. Just identify the scheduler you want to run using `schedctl list` and then operate it using `schedctl run` and `schedctl stop`.

Simple as that. The tool will take care of downloading the scheduler and start the binary inside it.

**Since containerized schedulers require extended capabilities, it's very likely that you'll need to run `schedctl` as root.**

## Development

`schedctl` is just a regular Go project, so just install `make`, `go` and you should be pretty much convered.

However, we have a couple things in the development workflow that are worth exploring.

### QEMU rootfs

We decided to run a significant portion of our integration tests in a QEMU virtual machine orchestrated by a testing library.

This means that we ship a pre-built QEMU rootfs to avoid rebuilding the root filesystem every time tests run, and this rootfs may occasionally need a refresh. In order to do so since the rootfs is based on an Arch Linux filesystem we have an Arch Linux distrobox ready to use.

To refresh the disk image:

```sh
$ distrobox assemble create --file testdata/distrobox.ini
$ distrobox enter arch-bootstrap
$ cd testdata
$ ./prepare_disk_image.sh
```

Since the rootfs versioning is managed through `git-lfs` you might want to run a `git rm rootfs.raw` before doing all of this.

### QEMU kernel image

We also ship a pre-built kernel for tests. The configuration is in [testdata/config](testdata/config).

To refresh the kernel image and build a new one, follow these steps:

```sh
$ distrobox assemble create --file testdata/distrobox.ini
$ distrobox enter arch-bootstrap
$ cd testdata
$ ./prepare_kernel_image.sh
```
