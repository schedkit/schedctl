#!/bin/bash

# Set variables
kernel_version=$1 # Kernel version tag as the first argument

# Check if kernel version is provided
if [ -z "$kernel_version" ]; then
 echo "Usage: $0 <kernel_version>"
 echo "Example: $0 v6.13"
 exit 1
fi

echo "Getting kernel source for version $kernel_version"
git clone --depth 1 --branch "$kernel_version" https://github.com/torvalds/linux.git kernel_source

echo "Configuring kernel"
cd kernel_source
make x86_64_defconfig
make kvm_guest.config

echo "Building kernel"
make -j$(nproc) bzImage

cp arch/x86/boot/bzImage ../

echo "Kernel compilation complete. Enjoy!"
