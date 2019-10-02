#!/usr/bin/env bash

set -eo pipefail

VERSION="$(./.gitlab/ci/scripts/version.sh 2>/dev/null || echo 'dev')"

# Generate Index page.
go run ./.gitlab/ci/scripts/generate-index-file/main.go bin/ ${VERSION} ${CI_COMMIT_REF_NAME} ${CI_COMMIT_SHA}

echo "Generated Index page"

aws s3 sync bin ${S3_URL} --acl public-read

# Copy the binaries to the latest directory.
LATEST_STABLE_TAG=$(git -c versionsort.prereleaseSuffix="-rc" tag -l "v*.*.*" --sort=-v:refname | awk '!/rc/' | head -n 1)
if [ $(git describe --exact-match --match ${LATEST_STABLE_TAG} >/dev/null 2>&1) ]; then
      aws s3 sync bin s3://${S3_BUCKET}/latest --acl public-read
fi

# Add assets to release page
bash ./.gitlab/ci/scripts/gitlab_release.sh
