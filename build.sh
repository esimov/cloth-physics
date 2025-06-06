#!/bin/bash
set -e

VERSION="1.0.4"

cd $(dirname "${BASH_SOURCE[0]}")
OD="$(pwd)"
WD=$OD

package() {
	echo Packaging $1 Binary
	bdir=cloth-physics-${VERSION}-$2-$3
	rm -rf packages/$bdir && mkdir -p packages/$bdir

	if [ "$2" == "windows" ]; then
		USE_WINDOWS_GUI_MODE="-H=windowsgui"
	else 
		USE_WINDOWS_GUI_MODE=""
	fi

	GOOS=$2 GOARCH=$3 go build -ldflags "$USE_WINDOWS_GUI_MODE -X main.Version=$VERSION" -o "$OD/cloth-physics" main.go

	if [ "$2" == "windows" ]; then
		mv cloth-physics packages/$bdir/cloth-physics.exe
	else
		mv cloth-physics packages/$bdir
	fi
	cp README.md packages/$bdir
	cd packages
	if [ "$2" == "linux" ]; then
		tar -zcf $bdir.tar.gz $bdir
	else
		zip -r -q $bdir.zip $bdir
	fi
	rm -rf $bdir
	cd ..
}

if [ "$1" == "package" ]; then
	rm -rf packages/
	package "Windows" "windows" "amd64"
	package "Mac" "darwin" "amd64"
	package "Linux" "linux" "amd64"
	exit
fi

# temp directory for storing isolated environment.
TMP="$(mktemp -d -t sdb.XXXX)"
rmtemp() {
	rm -rf "$TMP"
}
trap rmtemp EXIT

if [ "$NOCOPY" != "1" ]; then
	# copy all files to an isolated directory.
	WD="$TMP/src/github.com/esimov/cloth-physics"
	export GOPATH="$TMP"
	for file in `find . -type f`; do
		# TODO: use .gitignore to ignore, or possibly just use git to determine the file list.
		if [[ "$file" != "." && "$file" != ./.git* && "$file" != ./cloth-physics ]]; then
			mkdir -p "$WD/$(dirname "${file}")"
			cp -P "$file" "$WD/$(dirname "${file}")"
		fi
	done
	cd $WD
fi

