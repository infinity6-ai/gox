#!/bin/bash -e

eval "$(i6dev meta debug i6-go-build I6DEV_DEBUG)"

function cmd_clean() {
  local _k=""
  find -name target | while read _k; do
    rm -rf "$_k"
  done
}

function cmd_update() {
  i6dev golang update
}

function cmd_run() {
  go run main.go "$@"
}

function cmd_codegen() {
  go generate ./...
}

function cmd_run_box() {
  eval "$(i6dev box-config get-envs I6_BOX_COMMON_CONFIG)"
  cmd_run "$@"
}

function cmd_test_unit() {
  i6dev golang test_unit "$@"
  #strace go test -run '^TestUnit.*$' ./...
}

function cmd_test() {
  cmd_test_unit
}


function cmd_test_remote() {
  i6dev golang test_remote "$@"
}

function cmd_test_all() {
  i6dev golang test_all "$@"
}

function cmd_fmt() {
  go fmt ./...
}

function cmd_release() {
  i6dev golang bin-compile
  i6dev golang release
  i6dev golang bin-release latest
}

function cmd_version-inc {
  i6dev dever version-inc version/version.go
}

function cmd_version {
  go run version/main/main.go
}


function cmd_ds_index_create() {
  I6_GCP_PROJECT="i6-core-prod" i6dev gcp gcloud datastore indexes create -q --database=bla1 "sample/index.yaml"
}

cd "$(dirname "$0")"; _cmd="${1?"cmd is required"}"; shift; "cmd_${_cmd}" "$@"

