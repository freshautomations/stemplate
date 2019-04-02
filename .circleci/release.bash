#!/bin/bash

set -euo pipefail
# Opinionated script to release on GitHub.
# This script runs in CircleCI, in a golang docker container from a folder that is a git repo.
# The script expects the binaries to reside in the build folder.
STOML_VERSION=0.3.0

export TAG="`git tag --list --sort=v:refname --points-at HEAD | tail -1`"

echo "TAG=$TAG"
if [[ -z "$TAG" ]]; then
  echo "No tag to build."
  exit 1
fi

# Create GitHub release draft
draftdata="
{
  \"tag_name\": \"$TAG\",
  \"target_commitish\": \"master\",
  \"name\": \"$TAG\",
  \"body\": \"\",
  \"draft\": true,
  \"prerelease\": false
}
"
curl -s -S -X POST -u "${GITHUB_USERNAME}:${GITHUB_TOKEN}" https://api.github.com/repos/freshautomations/stemplate/releases --user-agent freshautomations -H "Accept: application/vnd.github.v3.json" -d "$draftdata" > draft.json
ERR=$?
if [[ $ERR -ne 0 ]]; then
  echo "ERROR: curl error, exitcode $ERR."
  exit $ERR
fi

#The proof of the pudding is the eating. - Cervantes
export id="`build/stoml_linux_amd64 draft.json id`"
if [ -z "$id" ]; then
  echo "ERROR: Could not get draft id."
  exit 1
fi

echo "Release ID: ${id}"

# Upload binaries

for binary in stemplate_darwin_386 stemplate_darwin_amd64 stemplate_linux_386 stemplate_linux_amd64 stemplate_windows_386 stemplate_windows_amd64
do
echo -ne "Processing ${binary}... "
if [[ ! -f "build/${binary}" ]]; then
  echo "${binary} does not exist."
  continue
fi
curl -s -S -X POST -u "${GITHUB_USERNAME}:${GITHUB_TOKEN}" "https://uploads.github.com/repos/freshautomations/stemplate/releases/${id}/assets?name=${binary}" --user-agent freshautomations -H "Accept: application/vnd.github.v3.raw+json" -H "Content-Type: application/octet-stream" -H "Content-Encoding: utf8" --data-binary "@build/${binary}" > upload.json
ERR=$?
if [[ $ERR -ne 0 ]]; then
  echo "ERROR: curl error, exitcode $ERR."
  exit $ERR
fi

wget -q "https://github.com/freshautomations/stoml/releases/download/v${STOML_VERSION}/stoml_linux_amd64"
chmod +x ./stoml_linux_amd64
export uid="`./stoml_linux_amd64 upload.json id`"
if [ -z "$uid" ]; then
  echo "ERROR: Could not get upload id for binary ${binary}."
  exit 1
fi

echo "uploaded binary ${binary}, id ${uid}."
done

rm draft.json
rm upload.json

# Publish release
releasedata="
{
  \"draft\": false,
  \"tag_name\": \"$TAG\"
}
"
curl -s -S -X POST -u "${GITHUB_USERNAME}:${GITHUB_TOKEN}" "https://api.github.com/repos/freshautomations/stemplate/releases/${id}" --user-agent script -H "Accept: application/vnd.github.v3.json" -d "$releasedata"
ERR=$?
if [[ $ERR -ne 0 ]]; then
  echo "ERROR: curl error, exitcode $ERR."
  exit $ERR
fi

