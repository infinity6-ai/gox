#!/bin/bash -xe

function cmd_build() {
  ./meta/ci/ci.sh build
  mkdir -p gen
  if ! i6dev code is-clean; then
    touch gen/dever.flag
    ./build.sh version-inc
  fi
}

function cmd_test() {
  ./meta/ci/ci.sh test
}

function cmd_release() {
  ./meta/ci/ci.sh release
}

cd "$(dirname "$0")/../.."; _cmd="${1?"cmd is required"}"; shift; "cmd_${_cmd}" "$@"
