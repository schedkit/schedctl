#!/bin/sh

dd if=/dev/zero of=rootfs.raw bs=1536MB count=1
mkfs.ext4 rootfs.raw
sudo losetup -fP rootfs.raw
mkdir rootfs
sudo mount /dev/loop0 rootfs
sudo pacstrap -c rootfs base openssh containerd nerdctl podman cni-plugins

echo "[Match]
Name=enp0s3

[Network]
DHCP=yes" | sudo tee rootfs/etc/systemd/network/20-wired.network

echo "nameserver 1.1.1.1
nameserver 8.8.8.8" | sudo tee rootfs/etc/resolv.conf

sudo sed -i '/^root/ { s/:x:/::/ }' rootfs/etc/passwd
sudo sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' rootfs/etc/ssh/sshd_config
sudo sed -i 's/#PermitEmptyPasswords no/PermitEmptyPasswords yes/' rootfs/etc/ssh/sshd_config

sudo arch-chroot rootfs systemctl enable sshd systemd-networkd containerd
# sudo rm rootfs/var/cache/pacman/pkg/*
sudo umount rootfs
sudo losetup -d /dev/loop0
rm -r rootfs

# Ensure the raw image is not writable
qemu-img create -o backing_file=rootfs.raw,backing_fmt=raw -f qcow2 rootfs.cow
