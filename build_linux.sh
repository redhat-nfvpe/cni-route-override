#!/usr/bin/env bash
set -e
cd $(dirname "$0")

if [ "$(uname)" == "Darwin" ]; then
	export GOOS="${GOOS:-linux}"
fi

export GOFLAGS="${GOFLAGS} -mod=vendor"

mkdir -p "${PWD}/bin"

echo "Building plugins ${GOOS}"
PLUGINS="cmd/route-override"
for d in $PLUGINS; do
	if [ -d "$d" ]; then
		plugin="$(basename "$d")"
		if [ $plugin != "windows" ]; then
			echo "  $plugin"
			${GO:-go} build -o "${PWD}/bin/$plugin" "$@" ./"$d"
		fi
	fi
done
