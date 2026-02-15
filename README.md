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

We run a significant portion of our integration tests in a QEMU virtual machine orchestrated by a testing library.

The QEMU rootfs is an Arch Linux disk image built with [mkosi](https://github.com/systemd/mkosi). The configuration lives in `testdata/` (`mkosi.conf`, `mkosi.repart/`, `mkosi.extra/`, and `mkosi.postinst.chroot`).

In CI, the image is built on-the-fly by mkosi. To build it locally:

```sh
sudo mkosi --directory testdata --output-dir testdata
```

This produces `testdata/rootfs.raw`. Note that mkosi creates a GPT-partitioned disk image, so you'll need to extract the root partition for use with the test framework:

```sh
LOOP=$(sudo losetup --find --show --partscan testdata/rootfs.raw)
sudo dd if="${LOOP}p1" of=testdata/rootfs_ext4.raw bs=4M
sudo losetup -d "$LOOP"
mv testdata/rootfs_ext4.raw testdata/rootfs.raw
qemu-img create -o backing_file=rootfs.raw,backing_fmt=raw -f qcow2 testdata/rootfs.cow
```

#### Legacy method

The old `testdata/prepare_disk_image.sh` script can still be used with distrobox to build the image manually:

```sh
$ distrobox assemble create --file testdata/distrobox.ini
$ distrobox enter arch-bootstrap
$ cd testdata
$ ./prepare_disk_image.sh
```

### QEMU kernel image

We also ship a pre-built kernel for tests. The configuration is in [testdata/config](testdata/config).

To refresh the kernel image and build a new one, follow these steps:

```sh
$ distrobox assemble create --file testdata/distrobox.ini
$ distrobox enter arch-bootstrap
$ cd testdata
$ ./prepare_kernel_image.sh
```
