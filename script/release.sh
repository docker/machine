#!/bin/bash

#Put your github username here, while testing performing new releases
#GITHUB_USER=jeanlaurent
GITHUB_USER=docker
GITHUB_REPO=machine

function usage {
  echo "Usage: "
  echo "   GITHUB_TOKEN=XXXXX release/release.sh 0.5.x"
}

function display {
  echo "🐳  $1"
  echo ""
}

function checkError {
  if [[ "$?" -ne 0 ]]; then
    echo "😡   $1"
    if [[ "$2" == "showUsage" ]]; then
        usage
    fi
    exit 1
  fi
}

function createMachine {
  docker-machine rm -f release 2> /dev/null
  docker-machine create -d virtualbox --virtualbox-cpu-count=2 --virtualbox-memory=2048 release
}

VERSION=$1
GITHUB_VERSION="v${VERSION}"
PROJECT_URL="git@github.com:${GITHUB_USER}/${GITHUB_REPO}"

RELEASE_DIR="$(git rev-parse --show-toplevel)/../release-${VERSION}"
GITHUB_RELEASE_FILE="github-release-${VERSION}.md"

if [[ -z "${VERSION}" ]]; then
  #TODO: Check version is well formed
  echo "Missing version argument"
  usage
  exit 1
fi

if [[ -z "${GITHUB_TOKEN}" ]]; then
  echo "GITHUB_TOKEN missing"
  usage
  exit 1
fi

command -v git > /dev/null 2>&1
checkError "You obviously need git, please consider installing it..." "showUsage"

command -v github-release > /dev/null 2>&1
checkError "github-release is not installed, go get -u github.com/aktau/github-release or check https://github.com/aktau/github-release, aborting." "showUsage"

command -v openssl > /dev/null 2>&1
checkError "You need openssl to generate binaries signature, brew install it, aborting." "showUsage"

command -v docker-machine > /dev/null 2>&1
checkError "You must have a docker-machine in your path" "showUsage"

LAST_RELEASE_VERSION=$(git describe --abbrev=0 --tags)
#TODO: ABORT if not found (very unlikely but could happen if two tags are on the same commits )
# this tag search is done on master not on the clone...

display "Starting release from ${LAST_RELEASE_VERSION} to ${GITHUB_VERSION} on ${PROJECT_URL} with token ${GITHUB_TOKEN}"
while true; do
    read -p "🐳  Do you want to proceed with this release? (y/n) > " yn
    echo ""
    case $yn in
        [Yy]* ) break;;
        [Nn]* ) exit;;
        * ) echo "😡   Please answer yes or no.";;
    esac
done

display "Checking machine 'release' status"
MACHINE_STATUS=$(docker-machine status release)
if [[ "$?" -ne 0 ]]; then
  display "Machine 'release' does not exist, creating it"
  createMachine
else
  if [[ "${MACHINE_STATUS}" != "Running" ]]; then
    display "Machine 'release' is not running, trying to start it."
    docker-machine start release
    if [[ "$?" -ne 0 ]]; then
      display "Machine 'release' could not be started, removing and creating a fresh new one."
      createMachine
    fi
    display "Loosing 5 seconds to the virtualbox gods."
    sleep 5
  fi
fi

eval $(docker-machine env release)
checkError "Machine 'release' is in a weird state, aborting."

if [[ -d "${RELEASE_DIR}" ]]; then
  display "Cleaning up ${RELEASE_DIR}"
  rm -rdf "${RELEASE_DIR}"
  checkError "Can't clean up ${RELEASE_DIR}. You should do it manually and retry."
fi

display "Cloning into ${RELEASE_DIR} from ${PROJECT_URL}"

mkdir -p "${RELEASE_DIR}"
checkError "Can't create ${RELEASE_DIR}, aborting."
git clone -q "${PROJECT_URL}" "${RELEASE_DIR}"
checkError "Can't clone into ${RELEASE_DIR}, aborting."

cd "${RELEASE_DIR}"

display "Bump version number to ${VERSION}"
#TODO: This only works with the version written in the version.go file
sed -i.bak s/"${VERSION}-dev"/"${VERSION}"/g version/version.go
checkError "Sed borkage..., aborting."

git add version/version.go
git commit -q -m"Bump version to ${VERSION}" -s
checkError "Can't git commit the version upgrade, aborting."
rm version/version.go.bak

display "Building in-container style"
USE_CONTAINER=true make clean validate build-x
checkError "Build error, aborting."

display "Generating github release"
cp -f script/release/github-release-template.md "${GITHUB_RELEASE_FILE}"
checkError "Can't find github release template"
CONTRIBUTORS=$(git log "${LAST_RELEASE_VERSION}".. --format="%aN" --reverse | sort | uniq | awk '{printf "- %s\n", $0 }')
CHANGELOG=$(git log "${LAST_RELEASE_VERSION}".. --oneline)

CHECKSUM=""
cd bin/
for file in $(ls docker-machine*); do
  SHA256=$(openssl dgst -sha256 < "${file}")
  MD5=$(openssl dgst -md5 < "${file}")
  LINE=$(printf "\n * **%s**\n  * sha256 \`%s\`\n  * md5 \`%s\`\n\n" "${file}" "${SHA256}" "${MD5}")
  CHECKSUM="${CHECKSUM}${LINE}"
done
cd ..

TEMPLATE=$(cat "${GITHUB_RELEASE_FILE}")
echo "${TEMPLATE//\{\{VERSION\}\}/$GITHUB_VERSION}" > "${GITHUB_RELEASE_FILE}"
checkError "Couldn't replace [ ${GITHUB_VERSION} ]"

TEMPLATE=$(cat "${GITHUB_RELEASE_FILE}")
echo "${TEMPLATE//\{\{CHANGELOG\}\}/$CHANGELOG}" > "${GITHUB_RELEASE_FILE}"
checkError "Couldn't replace [ ${CHANGELOG} ]"

TEMPLATE=$(cat "${GITHUB_RELEASE_FILE}")
echo "${TEMPLATE//\{\{CONTRIBUTORS\}\}/$CONTRIBUTORS}" > "${GITHUB_RELEASE_FILE}"
checkError "Couldn't replace [ ${CONTRIBUTORS} ]"

TEMPLATE=$(cat "${GITHUB_RELEASE_FILE}")
echo "${TEMPLATE//\{\{CHECKSUM\}\}/$CHECKSUM}" > "${GITHUB_RELEASE_FILE}"
checkError "Couldn't replace [ ${CHECKSUM} ]"

RELEASE_DOCUMENTATION="$(cat ${GITHUB_RELEASE_FILE})"

display "Tagging and pushing tags"
git remote | grep -q remote.prod.url
if [[ "$?" -ne 0 ]]; then
  display "Adding 'remote.prod.url' remote git url"
  git remote add remote.prod.url "${PROJECT_URL}"
fi

display "Checking if remote tag ${GITHUB_VERSION} already exists."
git ls-remote --tags 2> /dev/null | grep -q "${GITHUB_VERSION}" # returns 0 if found, 1 if not
if [[ "$?" -ne 1 ]]; then
  display "Deleting previous tag ${GITHUB_VERSION}"
  git tag -d "${GITHUB_VERSION}" &> /dev/null
  git push -q origin :refs/tags/"${GITHUB_VERSION}"
else
  echo "Tag ${GITHUB_VERSION} does not exist... yet"
fi

display "Tagging release on github"
git tag "${GITHUB_VERSION}"
git push -q remote.prod.url "${GITHUB_VERSION}"
checkError "Could not push to remote url"

display "Checking if release already exists"
github-release info \
    --security-token  "${GITHUB_TOKEN}" \
    --user "${GITHUB_USER}" \
    --repo "${GITHUB_REPO}" \
    --tag "${GITHUB_VERSION}" > /dev/null 2>&1

if [[ "$?" -ne 1 ]]; then
  display "Release already exists, cleaning it up."
  github-release delete \
      --security-token  "${GITHUB_TOKEN}" \
      --user "${GITHUB_USER}" \
      --repo "${GITHUB_REPO}" \
      --tag "${GITHUB_VERSION}"
  checkError "Could not delete release, aborting."
fi

display "Creating release on github"
github-release release \
    --security-token  "${GITHUB_TOKEN}" \
    --user "${GITHUB_USER}" \
    --repo "${GITHUB_REPO}" \
    --tag "${GITHUB_VERSION}" \
    --name "${GITHUB_VERSION}" \
    --description "${RELEASE_DOCUMENTATION}" \
    --pre-release
checkError "Could not create release, aborting."


display "Uploading binaries"
cd bin/
for file in $(ls docker-machine*); do
  display "Uploading ${file}..."
  github-release upload \
      --security-token  "${GITHUB_TOKEN}" \
      --user "${GITHUB_USER}" \
      --repo "${GITHUB_REPO}" \
      --tag "${GITHUB_VERSION}" \
      --name "${file}" \
      --file "${file}"
  if [[ "$?" -ne 0 ]]; then
    display "Could not upload ${file}, continuing with others."
  fi
done
cd ..

git remote rm remote.prod.url

rm ${GITHUB_RELEASE_FILE}

echo "There is a couple of tasks your still need to do manually."
echo "  1 Open the release notes created for you on github https://github.com/docker/machine/releases/tag/${GITHUB_VERSION}, you'll have a chance to enhance commit details a bit."
echo "  2 Once you're happy with your release notes on github, copy the list of changes to the CHANGELOG.md"
echo "  3 Update the documentation branch"
echo "  4 Test the binaries linked from the github release page"
echo "  6 Change version/version.go to the next version"
echo "  7 Party !!"
echo " The full details of these tasks are described in the RELEASE.md document, available at https://github.com/docker/machine/blob/master/docs/RELEASE.md"
