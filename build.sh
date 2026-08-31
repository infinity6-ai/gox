#!/bin/bash -e

eval "$(i6dev meta debug i6gox-build I6DEV_DEBUG 1>/dev/null 2>&1 || true)"

function cmd_comp_list() {
  find . -maxdepth 2 -name go.mod | cut -d'/' -f2
}

function cmd_comp_create() {
  local _name="${1?'_comp_name'}"
  mkdir -p "$_name"
  if [ ! -f "$_name/go.mod" ]; then
    echo "module github.com/infinity6-ai/gox/$_name" > "$_name/go.mod"
    echo "" >> "$_name/go.mod"
    echo "go $(i6dev util env I6_GO_VERSION)" >> "$_name/go.mod"
  fi
}

function cmd_comps_run() {
  local _k=""
  cmd_comp_list | while read _k; do
    ./comp.sh "$_k" "$@"
  done
}

function cmd_clean() {
  cmd_comps_run clean "$@"
}

function cmd_update() {
  # i6dev golang auth
  # I6DEV_GOLANG_AUTH_DISABLED="true" cmd_comps_run update "$@"
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

