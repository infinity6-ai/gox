#!/bin/bash -e

eval "$(i6dev meta debug i6gox-build I6DEV_DEBUG)"

function _run() {
  local _k=""
  ./list.sh | while read _k; do
    ./comp.sh "$_k" "$@"
  done
}

cd "$(dirname "$0")"
_run "$@"


