#!/bin/bash -e

eval "$(i6dev meta debug i6gox-comp I6DEV_DEBUG)"

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

function cmd_test_unit() {
  i6dev golang test_unit "$@"
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

_cmd="${1?"cmd is required"}"; shift; 
_comp="${2?"comp is required"}"; shift; 
cd "$(dirname "$0")/$_comp"; 
"cmd_${_cmd}" "$@"

