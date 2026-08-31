#!/bin/bash -e

[ "$I6DEV_DEBUG" != "true" ] || eval "$(i6dev meta debug i6gox-build I6DEV_DEBUG)"

function cmd_comp_list() {
  find . -maxdepth 2 -name go.mod | cut -d'/' -f2
}

function cmd_comps_run() {
  local _k=""
  cmd_comp_list | while read _k; do
    ./comp.sh "$_k" "$@"
  done
}

function cmd_work_init() {
  cmd_comps_run | xargs go work init
}

function cmd_clean() {
  cmd_comps_run clean "$@"
}

function cmd_update() {
  cmd_comps_run update "$@"
}

function cmd_codegen() {
  cmd_comps_run codegen "$@"
}

function cmd_test() {
  cmd_comps_run test "$@"
}

function cmd_fmt() {
  cmd_comps_run fmt "$@"
}

function cmd_release() {
  cmd_comps_run release
}

cd "$(dirname "$0")"; _cmd="${1?"cmd is required"}"; shift; "cmd_${_cmd}" "$@"

