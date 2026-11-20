#!/bin/bash -e

TMP_DIR="$(mktemp -d)"
BUILD_DIR="$(mktemp -d)"
BIN="$BUILD_DIR/brwjsz"

cleanup() {
  rm -rf "$TMP_DIR" "$BUILD_DIR"
}
trap cleanup EXIT

ensure_binary() {
  if [ ! -f "$BIN" ]; then
    echo "==> Building brwjsz binary..."
    go build -o "$BIN" .
  fi
}

assert_eq() {
  local expected="$1"
  local actual="$2"
  local msg="${3:-Assertion failed}"
  if [ "$expected" != "$actual" ]; then
    echo "ERROR: $msg"
    echo "  Expected: $expected"
    echo "  Actual:   $actual"
    exit 1
  fi
}

function cmd_version() {
  ensure_binary
  echo "--- Testing: version ---"
  local ver
  ver="$($BIN version)"
  [ -n "$ver" ] || { echo "version output is empty"; exit 1; }
  echo "  OK: version=$ver"
}

function cmd_exec_inline() {
  ensure_binary
  echo "--- Testing: exec inline expression ---"
  local res
  res="$($BIN exec "21 * 2")"
  assert_eq "42" "$res" "inline eval 21 * 2"
  echo "  OK: exec inline expression"
}

function cmd_exec_code_flag() {
  ensure_binary
  echo "--- Testing: exec --code flag ---"
  local res
  res="$($BIN exec --code "10 + 15")"
  assert_eq "25" "$res" "code flag 10 + 15"
  echo "  OK: exec --code flag"
}

function cmd_exec_file() {
  ensure_binary
  echo "--- Testing: exec --file ---"
  local script_file="$TMP_DIR/script.js"
  echo "const a = 12; const b = 3; a / b;" > "$script_file"
  local res
  res="$($BIN exec --file "$script_file")"
  assert_eq "4" "$res" "exec file"
  echo "  OK: exec --file"
}

function cmd_exec_stdin() {
  ensure_binary
  echo "--- Testing: exec stdin ---"
  local res
  res="$(echo "7 * 6" | $BIN exec --file -)"
  assert_eq "42" "$res" "exec stdin"
  echo "  OK: exec stdin"
}

function cmd_exec_args() {
  ensure_binary
  echo "--- Testing: exec with function and multiple --arg ---"
  local res
  res="$($BIN exec --code "(a, b) => a + b" --arg 10 --arg 25)"
  assert_eq "35" "$res" "function arguments sum"
  echo "  OK: exec function with multiple args"
}

function cmd_exec_arg_json() {
  ensure_binary
  echo "--- Testing: exec with JSON object --arg ---"
  local res
  res="$($BIN exec --code "(obj) => obj.multiplier * 3" --arg '{"multiplier": 7}')"
  assert_eq "21" "$res" "json argument member access"
  echo "  OK: exec with JSON object arg"
}

function cmd_exec_arg_file() {
  ensure_binary
  echo "--- Testing: exec with --arg-file and --raw ---"
  local arg_file="$TMP_DIR/user.json"
  echo '{"name": "Alice", "role": "admin"}' > "$arg_file"
  local res
  res="$($BIN exec --code "(user) => user.name" --arg-file "$arg_file" --raw)"
  assert_eq "Alice" "$res" "arg-file parsing"
  echo "  OK: exec with --arg-file and --raw"
}

function cmd_exec_async() {
  ensure_binary
  echo "--- Testing: exec async function ---"
  local res
  res="$($BIN exec --code "async () => { await new Promise(r => setTimeout(r, 20)); return 'completed'; }" --raw)"
  assert_eq "completed" "$res" "async function awaiting"
  echo "  OK: exec async function"
}

function cmd_exec_site() {
  ensure_binary
  echo "--- Testing: exec with --site ---"
  local site_dir="$TMP_DIR/site"
  mkdir -p "$site_dir"
  echo '{"title": "awesome site data"}' > "$site_dir/app.json"
  local res
  res="$($BIN exec --site "$site_dir" --code "async () => { const r = await fetch('/app.json'); return (await r.json()).title; }" --raw)"
  assert_eq "awesome site data" "$res" "fetch from disk site"
  echo "  OK: exec with --site"
}

function cmd_exec_error() {
  ensure_binary
  echo "--- Testing: exec error handling ---"
  if $BIN exec --code "throw new Error('deliberate failure')" > /dev/null 2>&1; then
    echo "ERROR: exec should have failed on thrown error"
    exit 1
  fi
  echo "  OK: exec error handling"
}

function cmd_server_lifecycle() {
  ensure_binary
  echo "--- Testing: server start/exec/stop lifecycle ---"
  local state_file="$TMP_DIR/server.state"

  # 1. Start detached daemon
  $BIN server start --state "$state_file" --detach
  [ -f "$state_file" ] || { echo "ERROR: state file not created"; exit 1; }

  # 2. Exec inline code on running daemon
  local res1
  res1="$($BIN exec --state "$state_file" --code "21 * 2")"
  assert_eq "42" "$res1" "server exec 21 * 2"

  # 3. Exec with function and arguments
  local res2
  res2="$($BIN exec --state "$state_file" --code "(a, b) => a + b" --arg 10 --arg 20)"
  assert_eq "30" "$res2" "server exec (a, b) => a + b"

  # 4. State persistence across calls: set global in call 1, read it in call 2
  $BIN exec --state "$state_file" --code "window.myCounter = 100" > /dev/null
  local res3
  res3="$($BIN exec --state "$state_file" --code "window.myCounter += 25; window.myCounter")"
  assert_eq "125" "$res3" "server state persistence between calls"

  # 5. Stop the daemon
  $BIN server stop --state "$state_file"
  [ ! -f "$state_file" ] || { echo "ERROR: state file still exists after stop"; exit 1; }

  echo "  OK: server start/exec/stop lifecycle"
}

function cmd_all() {
  echo "=========================================="
  echo "Running all brwjsz CLI command tests"
  echo "=========================================="
  cmd_version
  cmd_exec_inline
  cmd_exec_code_flag
  cmd_exec_file
  cmd_exec_stdin
  cmd_exec_args
  cmd_exec_arg_json
  cmd_exec_arg_file
  cmd_exec_site
  cmd_exec_async
  cmd_exec_error
  cmd_server_lifecycle
  echo "=========================================="
  echo "All brwjsz CLI tests passed successfully!"
  echo "=========================================="
}

cd "$(dirname "$0")"; _cmd="${1?"cmd is required"}"; shift; "cmd_${_cmd}" "$@"
