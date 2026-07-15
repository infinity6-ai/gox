#!/bin/bash -e

(
  cd "$(dirname "$0")" && \
  find . -maxdepth 2 -name go.mod | cut -d'/' -f2
)

