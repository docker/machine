#!/usr/bin/env bash

set -eo pipefail

VERSION="$(./.gitlab/ci/scripts/version.sh 2>/dev/null || echo 'dev')"

# Installing release index generator
RELEASE_INDEX_GEN_VERSION=${RELEASE_INDEX_GEN_VERSION:-latest}
releaseIndexGen=.tmp/release-index-gen-${RELEASE_INDEX_GEN_VERSION}
OS_TYPE=$(uname -s | tr '[:upper:]' '[:lower:]')
DOWNLOAD_URL="https://storage.googleapis.com/gitlab-runner-tools/release-index-generator/${RELEASE_INDEX_GEN_VERSION}/release-index-gen-${OS_TYPE}-amd64"

mkdir -p $(dirname ${releaseIndexGen})
curl -sL "${DOWNLOAD_URL}" -o "${releaseIndexGen}"
chmod +x "${releaseIndexGen}"

# Generate Index page
"${releaseIndexGen}" -working-directory bin/ \
                  -project-version ${VERSION} \
                  -project-git-ref ${CI_COMMIT_REF_NAME} \
                  -project-git-revision ${CI_COMMIT_SHA} \
                  -project-name "Docker Machine (GitLab's fork)" \
                  -project-repo-url "https://gitlab.com/gitlab-org/ci-cd/docker-machine" \
                  -gpg-key-env GPG_KEY \
                  -gpg-password-env GPG_PASSPHRASE

echo "Generated Index page"

aws s3 sync bin ${S3_URL} --acl public-read

# Copy the binaries to the latest directory.
LATEST_STABLE_TAG=$(git -c versionsort.prereleaseSuffix="-rc" tag -l "v*.*.*" --sort=-v:refname | awk '!/rc/' | head -n 1)
if [ $(git describe --exact-match --match ${LATEST_STABLE_TAG} >/dev/null 2>&1) ]; then
      aws s3 sync bin s3://${S3_BUCKET}/latest --acl public-read
fi

# Add assets to release page
bash ./.gitlab/ci/scripts/gitlab_release.sh
