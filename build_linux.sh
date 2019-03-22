#!/usr/bin/env bash
set -e
cd $(dirname "$0")

if [ "$(uname)" == "Darwin" ]; then
	export GOOS="${GOOS:-linux}"
fi

ORG_PATH="github.com/redhat-nfvpe"
export REPO_PATH="${ORG_PATH}/cni-route-overwrite"

if [ ! -h gopath/src/${REPO_PATH} ]; then
	mkdir -p gopath/src/${ORG_PATH}
	ln -s ../../../.. gopath/src/${REPO_PATH} || exit 255
fi

export GOPATH=${PWD}/gopath
export GO="${GO:-go}"

mkdir -p "${PWD}/bin"

echo "Building plugins ${GOOS}"
PLUGINS="cmd/route-overwrite"
for d in $PLUGINS; do
	if [ -d "$d" ]; then
		plugin="$(basename "$d")"
		if [ $plugin != "windows" ]; then
			echo "  $plugin"
			$GO build -o "${PWD}/bin/$plugin" "$@" "$REPO_PATH"/$d
		fi
	fi
done
