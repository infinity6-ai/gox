#!/bin/bash -e

eval "$(i6dev meta debug i6gox-build I6DEV_DEBUG)"

function cmd_comps_run() {
  local _k=""
  ./list.sh | while read _k; do
    ./comp.sh "$_k" "$@"
  done
}

function cmd_clean() {
  cmd_comps_run clean "$@"
}

function cmd_update() {
  i6dev golang auth
  I6DEV_GOLANG_AUTH_DISABLED="true" cmd_comps_run update "$@"
}

function cmd_codegen() {
  cmd_comps_run codegen "$@"
}

function cmd_test() {
  cmd_comps_run test "$@"
}

function cmd_test_remote() {
  cmd_comps_run test_remote "$@"
}

function cmd_test_all() {
  cmd_comps_run test_all "$@"
}

function cmd_fmt() {
  cmd_comps_run fmt "$@"
}

cd "$(dirname "$0")"; _cmd="${1?"cmd is required"}"; shift; "cmd_${_cmd}" "$@"

