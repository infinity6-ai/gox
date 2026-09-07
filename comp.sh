#!/bin/bash -e

[ "$I6DEV_DEBUG" != "true" ] || eval "$(i6dev meta debug i6gox-comp I6DEV_DEBUG)"

function cmd_clean() {
  local _k=""
  find -name target | while read _k; do
    rm -rf "$_k"
  done
}

function cmd_update() {
  go mod tidy
  cmd_codegen
}

function cmd_run() {
  go run . "$@"
}

function cmd_codegen() {
  go generate ./...
}

function cmd_test_unit() {
  go test -run '^TestUnit.*$' ./... "$@" 
}

function cmd_test_remote() {
  go test -run '^TestRemote.*$' ./... "$@" 
}

function cmd_test() {
  cmd_test_unit
}

function cmd_fmt() {
  go fmt ./...
}

function cmd_release() {
  [ -z "$(git status -s "$@")" ]
  [ "x0" == "x$(git rev-list --count @{u}..HEAD)" ]
  local _version="$(cmd_run version)"
  [ ! -z "$_version" ]
  local _tag="$_comp/$_version"
  if git show-ref --tags "$_tag" --quiet; then
    echo "tag already exists: $_tag" 1>&2
    false
  fi
  git tag "$_tag"
  git push origin "$_tag" 
}

# function _go_base_path() {
#   GOWORK=off go list -m | rev | cut -d'/' -f2- | rev
# }

# function cmd_set_version() {
#   local _version="${1?'_version'}"
#   local _go_base_path="$(_go_base_path)"
#   # GOWORK=off go list -mod=readonly -m -f '{{.Path}}' $_go_base_path/...
#   local _go_dep_mod=""
#   GOWORK=off go list -mod=readonly -m "$_go_base_path/..." | \
#     grep "^$_go_base_path/.*\ v" | \
#     cut -d' ' -f1 | while read _go_dep_mod; do
#     go mod edit -require="${_go_dep_mod}@${_version}"
#   done
# }

# function cmd_internal_update() {
#   local _version="${1?'_version'}"
#   local _go_base_path="$(_go_base_path)"
#   go get "$_go_base_path/...@$_version"
#   go mod tidy
# }

_comp="${1?"comp is required"}"; shift; 
_cmd="${1?"cmd is required"}"; shift; 
cd "$(dirname "$0")/$_comp"; 
"cmd_${_cmd}" "$@"

