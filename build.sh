#!/bin/bash -e

[ "$I6DEV_DEBUG" != "true" ] || eval "$(i6dev meta debug i6gox-build I6DEV_DEBUG)"

function cmd_comps_list() {
  # find . -maxdepth 2 -name go.mod | cut -d'/' -f2
  echo "versionz"
  # echo "commonz"
  # echo "cryptz"
  # echo "schemaz"
  # echo "httpz"
  # echo "routez"
  # echo "msgz"
  # echo "fsz" 
}

function cmd_comps_exec() {
  local _k=""
  cmd_comps_list | while read _k; do
    (cd "$_k" && "$@")
  done
}

function cmd_comps_run() {
  local _k=""
  cmd_comps_list | while read _k; do
    ./comp.sh "$_k" "$@"
  done
}

function cmd_work_init() {
  [ ! -f go.work.sum ] || rm go.work.sum
  [ ! -f go.work ] || rm go.work
  cmd_comps_list | xargs go work init
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

function cmd_force_delete_version() {
  local _version="${1?'_version'}"
  [ ! -z "$_version" ]
  cmd_comps_list | while read _k; do
    git tag -d "$_k/$_version" || true
    git push --delete origin "$_k/$_version" || true
  done
}

function cmd_release() {
  [ -z "$(git status -s "$@")" ]
  [ "x0" == "x$(git rev-list --count @{u}..HEAD)" ]
  GOWORK=off cmd_comps_run release
}

# function cmd_set_version() {
#   local _version="${1?'_version'}"
#   # cmd_comps_run set_version "${_version}"
#   # git tag "$_version"
#   # git push origin "$_version"
#   cmd_comps_run set_version "${_version}"
# }

# function cmd_force_delete_tag() {
#   local _version="${1?'_version'}"
#   git tag -d "$_version"
#   git push --delete origin "$_version"
# }

cd "$(dirname "$0")"; _cmd="${1?"cmd is required"}"; shift; "cmd_${_cmd}" "$@"

