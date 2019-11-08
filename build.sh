#!/usr/bin/env bash
set -e

cd $(dirname "$0")
CWD=$(pwd)

# Use this to ensure that we have all the tools required to do a build.
export GO111MODULE=on
export GOFLAGS="-mod=vendor"

MISSING=()

check() {
  local X=$1
  set +e
  command -v $X >/dev/null 2>&1
  local RESULT=$?
  set -e
  if [ $RESULT != 0 ]; then
    MISSING+=($X)
  fi
}

check jq
check go

if ! [ ${#MISSING[@]} -eq 0 ]; then
  echo "Missing prerequisites:"
  for X in $MISSING; do
    echo "  $X"
  done

  exit 1
fi

echo "Prerequisites present"

VERSION=$(cat .shipyard/manifest.json | jq -r '.version')
echo "Build version '$VERSION'"

cd "$CWD"

# default to mostly true, set env val to override
DO_TEST=${DO_TEST:-"true"}
DO_VERIFY=${DO_VERIFY:-"true"}
DO_VET=${DO_VET:-"true"}

while [[ $# -gt 0 ]]; do
  key="$1"
  case $key in
    --fast)
        DO_VET="false"
        DO_VERIFY="false"
        DO_TEST="false"
        shift
        ;;
    --no-test)
        DO_TEST="false"
        shift
        ;;
    --no-verify)
        DO_VERIFY="false"
        shift
        ;;
    --no-vet)
        DO_VET="false"
        shift
        ;;
    *)
      shift
      ;;
  esac
done

if [[ $DO_TEST == "true" ]]; then
  if ! [ -x $CWD/bin/mockgen ]; then
    echo "Building mockgen"
    time go build -o $CWD/bin/mockgen $CWD/vendor/github.com/golang/mock/mockgen
  fi
  echo "mockgen tool checked"

  # echo "Generating mock for Requester"
  # time $CWD/bin/mockgen -destination=$CWD/mocks/mock_request.go -package=mocks github.com/object88/devex/requests Requester

  echo "Generating mock for RoundTripper"
  time $CWD/bin/mockgen -destination=$CWD/mocks/mock_httproundtripper.go -package=mocks net/http RoundTripper
fi

if [[ $DO_VERIFY == "true" ]]; then
    echo "Verifying modules"
    # returns non-zero if this doesn't verify out
    time go mod verify
fi

if [[ $DO_VET == "true" ]]; then
  # Vet's exit code is non-zero for erroneous invocation of the tool
  # or if a problem was reported, and 0 otherwise. Note that the
  # tool does not check every possible problem and depends on
  # unreliable heuristics, so it should be used as guidance only,
  # not as a firm indicator of program correctness.
  # [snip]
  # By default, all checks are performed.
  #
  # https://golang.org/cmd/vet/
  echo "Running the vet"
  time go vet $(go list ./...)
fi

# test executables and binaries
if [[ $DO_TEST == "true" ]]; then
  time go test ./... -count=1
fi

# build executable(s)
# method found here https://www.digitalocean.com/community/tutorials/how-to-build-go-executables-for-multiple-platforms-on-ubuntu-16-04

DEFAULT_GOOS=$(uname | tr '[:upper:]' '[:lower:]')
PLATFORMS=( "$DEFAULT_GOOS/amd64" )
if [ "$BUILD_AND_RELEASE" == "true" ]; then
  PLATFORMS=( "linux/amd64" "darwin/amd64" )
fi

# build executable for each platform...
for PLATFORM in "${PLATFORMS[@]}"; do
  GOOS=$(cut -d'/' -f1 <<< $PLATFORM)
  GOARCH=$(cut -d'/' -f2 <<< $PLATFORM)
  BINARY_NAME="churl-${GOOS}-${GOARCH}"
  echo "building as $BINARY_NAME"

  if [ $(uname) == "Darwin" ]; then
    # Cannot do a static compilation on Darwin.
    time env GOOS=$GOOS GOARCH=$GOARCH go build -o ./bin/$BINARY_NAME -ldflags "-s -w -X github.com/object88/churl/churl.ChurlVersion=$VERSION" ./main/main.go
  else
    time env GOOS=$GOOS GOARCH=$GOARCH go build -o ./bin/$BINARY_NAME -tags "netgo" -ldflags "-s -w -extldflags \"-static\" -X github.com/object88/churl/churl.ChurlVersion=$VERSION" ./main/main.go
  fi
done