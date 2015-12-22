#!/bin/bash -e

if ! which gox > /dev/null; then
  echo "gox is not installed..."
  exit 1
fi

git_version=$(git describe)
if git_status=$(git status --porcelain 2>/dev/null) && [ -n "${git_status}" ]; then
  git_version="${git_version}-dirty"
fi

ldflags="-X github.com/vmware/govmomi/govc/version.gitVersion ${git_version}"
BUILD_OS=${BUILD_OS:-darwin linux windows freebsd}
BUILD_ARCH=${BUILD_ARCH:-386 amd64}

gox \
  -parallel=1 \
  -ldflags="${ldflags}" \
  -os="${BUILD_OS}" \
  -arch="${BUILD_ARCH}" \
  github.com/vmware/govmomi/govc
