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

## Usage

Starting and stopping a scheduler using schedctl is trivial. Just identify the scheduler you want to run using `schedctl list` and then operate it using `schedctl start` and `schedctl stop`.

Simple as that. The tool will take care of downloading the scheduler and start the binary inside it.
