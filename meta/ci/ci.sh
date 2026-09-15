#!/bin/bash -xe

function cmd_build() {
  ./build.sh update
}

function cmd_test() {
  ./build.sh test
}

function cmd_release() {
  ./build.sh release
}

cd "$(dirname "$0")/../.."; _cmd="${1?"cmd is required"}"; shift; "cmd_${_cmd}" "$@"
