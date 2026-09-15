#!/bin/bash -xe

function cmd_image_download() {
  docker pull us-central1-docker.pkg.dev/i6-rs-contint/i6devhub/i6dev-mini:latest
  docker tag us-central1-docker.pkg.dev/i6-rs-contint/i6devhub/i6dev-mini:latest infinity6/current:dev
}

function cmd_build() {
  ./build.sh update
}

function cmd_test() {
  ./build.sh test
}

function cmd_release() {
  # ./build.sh release
  true
}

cd "$(dirname "$0")/../.."; _cmd="${1?"cmd is required"}"; shift; "cmd_${_cmd}" "$@"
