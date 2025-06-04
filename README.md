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

Starting and stopping a scheduler using schedctl is trivial. Just identify the scheduler you want to run using `schedctl list` and then operate it using `schedctl start` and `schedctl stop`.

Simple as that. The tool will take care of downloading the scheduler and start the binary inside it.
