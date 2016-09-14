#!/bin/bash
#  This file is part of cc-oci-runtime.
#
#  Copyright (C) 2016 Intel Corporation
#
#  This program is free software; you can redistribute it and/or
#  modify it under the terms of the GNU General Public License
#  as published by the Free Software Foundation; either version 2
#  of the License, or (at your option) any later version.
#
#  This program is distributed in the hope that it will be useful,
#  but WITHOUT ANY WARRANTY; without even the implied warranty of
#  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
#  GNU General Public License for more details.
#
#  You should have received a copy of the GNU General Public License
#  along with this program; if not, write to the Free Software
#  Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301, USA.

set -e -x

cwd=$(cd `dirname "$0"`; pwd -P)

local_versions="$cwd/versions.txt"
cached_versions="/usr/local/lib/cached_versions.txt"

source $cwd/versions.txt

[ -e "${cached_versions}" ] && source "$cached_versions"

# Ensure "make install" as root can find clang
#
# See: https://github.com/travis-ci/travis-ci/issues/2607
export CC=$(which "$CC")

gnome_dl=https://download.gnome.org/sources

# Install required dependencies to build
# glib, json-glib, libmnl-dev check and cc-oci-runtime
sudo apt-get -qq install valgrind lcov uuid-dev pkg-config \
  zlib1g-dev libffi-dev gettext libpcre3-dev cppcheck \
  libmnl-dev libcap-ng-dev libgmp-dev libmpfr-dev libmpc-dev \
  libpixman-1-dev rpm2cpio

mkdir cor-dependencies
pushd cor-dependencies

# Build glib
if [ "${glib_version}" != "${cached_glib_version}" ]
then
    glib_major=`echo $glib_version | cut -d. -f1`
    glib_minor=`echo $glib_version | cut -d. -f2`
    curl -L -O "$gnome_dl/glib/${glib_major}.${glib_minor}/glib-${glib_version}.tar.xz"
    tar -xf "glib-${glib_version}.tar.xz"
    pushd "glib-${glib_version}"
    ./configure --disable-silent-rules
    make -j5
    sudo make install
    popd
fi

# Build json-glib
if [ "${json_glib_version}" != "${cached_json_glib_version}" ]
then
    json_major=`echo $json_glib_version | cut -d. -f1`
    json_minor=`echo $json_glib_version | cut -d. -f2`
    curl -L -O "$gnome_dl/json-glib/${json_major}.${json_minor}/json-glib-${json_glib_version}.tar.xz"
    tar -xf "json-glib-${json_glib_version}.tar.xz"
    pushd "json-glib-${json_glib_version}"
    ./configure --disable-silent-rules
    make -j5
    sudo make install
popd
fi

# Build check
# We need to build check as the check version in the OS used by travis isn't
# -pedantic safe.
if [ "${check_version}" != "${cached_check_version}" ]
then
    curl -L -O "https://github.com/libcheck/check/releases/download/${check_version}/check-${check_version}.tar.gz"
    tar -xf "check-${check_version}.tar.gz"
    pushd "check-${check_version}"
    ./configure
    make -j5
    sudo make install
    popd
fi

# Install bats
# (Note: no version, just check if the binary is there)
if [ ! -x /usr/local/bin/bats ]
then
    git clone https://github.com/sstephenson/bats.git
    pushd bats
    sudo ./install.sh /usr/local
    popd
fi

# build gcc (required for qemu-lite)
export PATH="/usr/local/gcc-${gcc_version}:$PATH"
if [ "${gcc_version}" != "${cached_gcc_version}" ]
then
    curl -L -O "http://mirrors.kernel.org/gnu/gcc/gcc-${gcc_version}/gcc-${gcc_version}.tar.bz2"
    tar xf "gcc-${gcc_version}.tar.bz2"
    pushd "gcc-${gcc_version}"
    ./configure --enable-languages=c --disable-multilib --prefix="/usr/local/gcc-${gcc_version}"
    make -j5
    sudo make install
    popd
fi

# build qemu-lite
if [ "${qemu_lite_version}" != "${cached_qemu_lite_version}" ]
then
    curl -L -O "https://github.com/01org/qemu-lite/archive/${qemu_lite_version}.tar.gz"
    tar xf "${qemu_lite_version}.tar.gz"
    pushd "qemu-lite-${qemu_lite_version}"
    CC="gcc" ./configure \
        --disable-bluez \
        --disable-brlapi \
        --disable-bzip2 \
        --disable-curl \
        --disable-curses \
        --disable-debug-tcg \
        --disable-fdt \
        --disable-glusterfs \
        --disable-gtk \
        --disable-libiscsi \
        --disable-libnfs \
        --disable-libssh2 \
        --disable-libusb \
        --disable-linux-aio \
        --disable-lzo \
        --disable-opengl \
        --disable-qom-cast-debug \
        --disable-rbd \
        --disable-rdma \
        --disable-sdl \
        --disable-seccomp \
        --disable-slirp \
        --disable-snappy \
        --disable-spice \
        --disable-strip \
        --disable-tcg-interpreter \
        --disable-tcmalloc \
        --disable-tools \
        --disable-tpm \
        --disable-usb-redir \
        --disable-uuid \
        --disable-vnc \
        --disable-vnc-{jpeg,png,sasl} \
        --disable-vte \
        --disable-xen \
        --enable-attr \
        --enable-cap-ng \
        --enable-kvm \
        --enable-virtfs \
        --target-list=x86_64-softmmu \
        --extra-cflags="-fno-semantic-interposition -O3 -falign-functions=32" \
        --prefix=/usr/local \
        --datadir=/usr/local/share/qemu-lite \
        --libdir=/usr/local/lib64/qemu-lite \
        --libexecdir=/usr/local/libexec/qemu-lite
    make -j5
    sudo make install
    popd
fi

# install kernel + Clear Containers image
mkdir artifacts
pushd artifacts
clr_release=$(curl -L https://download.clearlinux.org/latest)
clr_kernel_base_url="https://download.clearlinux.org/releases/${clr_release}/clear/x86_64/os/Packages"

clr_assets_dir=/usr/share/clear-containers/
sudo mkdir -p "$clr_assets_dir"

# find newest containers kernel
clr_kernel=$(curl -l -s -L "${clr_kernel_base_url}" |\
    grep -o "linux-container-[0-9][0-9.-]*\.x86_64.rpm" |\
    sort -u)

# download kernel
if [ ! -e "${clr_assets_dir}/${clr_kernel}" ]
then
    curl -L -O "${clr_kernel_base_url}/${clr_kernel}"

    # install kernel
    # (note: cpio on trusty does not support "-D")
    rpm2cpio "${clr_kernel}"| (cd / && sudo cpio -idv)
fi

# download image
if [ ! -e "${clr_assets_dir}/${clr_image}" ]
then
    clr_image_url="https://download.clearlinux.org/current/clear-${clr_release}-containers.img.xz"
    clr_image_compressed=$(basename "$clr_image_url")
    
    # uncompressed image name
    clr_image=${clr_image_compressed/.xz/}
    
    for file in "${clr_image_url}-SHA512SUMS" "${clr_image_url}"
    do
        curl -L -O "$file"
    done
    
    # verify image
    sha512sum -c "${clr_image_compressed}-SHA512SUMS"
    
    # unpack image
    unxz "${clr_image_compressed}"
    
    # install image
    sudo install "${clr_image}" "${clr_assets_dir}"
fi

# change kernel+image ownership
sudo chown -R $USER "${clr_assets_dir}"

# create image symlink (kernel will already have one)
clr_image_link=clear-containers.img
if [ ! -e "${clr_assets_dir}/${clr_image_link}" ]
then
    (cd "${clr_assets_dir}" && sudo ln -s "${clr_image}" "${clr_image_link}")
fi

mkdir docker
pushd docker

# create an ubuntu container
sudo docker run ubuntu true

# flatten the just-created ubuntu container
sudo docker export $(sudo docker ps -n 1 -q) | sudo docker import - cc-test-image

# export the container as a tar file
sudo docker save cc-test-image > cc-test-image.tar

# extra the container files
tar xvf cc-test-image.tar

# find the filesystem layer
# (only 1 layer as the image has been flattened)
fs=$(find . -name layer.tar)

# extract container image
mkdir -p /tmp/bundle/rootfs
sudo tar -C /tmp/bundle/rootfs -xvf "$fs"
popd

popd

# copy the "versions database" to the cache directory if it's not
# already there. Since /usr/local is our Travis cache directory, we can
# identify on the next build if these build dependencies need to
# rebuilt. Most of time they won't (but will if a version is bumped in
# the versions database).
sudo mkdir -p $(dirname "${cached_versions}")
egrep -v "^(\#|$)" "${local_versions}" |\
    sed 's/^/cached_/g' |\
    sudo tee "${cached_versions}"
sudo chmod 444 "${cached_versions}"
