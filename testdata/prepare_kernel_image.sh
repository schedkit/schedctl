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
git clone --depth 1 --branch "$kernel_version" git://git.kernel.org/pub/scm/linux/kernel/git/stable/linux-stable.git kernel_source

echo "Configuring kernel"
cd kernel_source
cp ../config .config

echo "Building kernel"
make -j$(nproc) bzImage

cp arch/x86/boot/bzImage ../

echo "Kernel compilation complete. Enjoy!"
