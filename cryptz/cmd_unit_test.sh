#!/bin/bash -e

# [ "$I6DEV_DEBUG" != "true" ] || eval "$(i6dev meta debug i6gox-comp I6DEV_DEBUG)"

TMP_DIR="$(mktemp -d)"
BUILD_DIR="$(mktemp -d)"
BIN="$BUILD_DIR/cryptz"

cleanup() {
  rm -rf "$TMP_DIR" "$BUILD_DIR"
}
trap cleanup EXIT

ensure_binary() {
  if [ ! -f "$BIN" ]; then
    echo "==> Building cryptz binary..."
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

function cmd_b64() {
  ensure_binary
  echo "--- Testing: b64 ---"
  local enc dec
  enc=$(echo -n "hello world" | $BIN b64 encode)
  assert_eq "aGVsbG8gd29ybGQ=" "$enc" "b64 encode"
  dec=$(echo -n "$enc" | $BIN b64 decode)
  assert_eq "hello world" "$dec" "b64 decode"

  local enc_url dec_url
  enc_url=$(echo -n "hello??world" | $BIN b64 encode --url)
  dec_url=$(echo -n "$enc_url" | $BIN b64 decode --url)
  assert_eq "hello??world" "$dec_url" "b64 url-safe roundtrip"
  echo "  OK: b64 encode/decode"
}

function cmd_b32() {
  ensure_binary
  echo "--- Testing: b32 ---"
  local enc dec
  enc=$(echo -n "hello world" | $BIN b32 encode)
  assert_eq "d1imor3f41rmusjccg" "$enc" "b32 encode"
  dec=$(echo -n "$enc" | $BIN b32 decode)
  assert_eq "hello world" "$dec" "b32 decode"
  echo "  OK: b32 encode/decode"
}

function cmd_hash() {
  ensure_binary
  echo "--- Testing: hash ---"
  local sha md5
  sha=$(echo -n "hello" | $BIN hash sha256)
  assert_eq "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824" "$sha" "hash sha256"
  md5=$(echo -n "hello" | $BIN hash md5)
  assert_eq "5d41402abc4b2a76b9719d911017c592" "$md5" "hash md5"
  echo "  OK: hash sha256 and md5"
}

function cmd_rand() {
  ensure_binary
  echo "--- Testing: rand ---"
  local hex b64 b64url
  hex=$($BIN rand -s 16)
  [ "${#hex}" -eq 32 ] || { echo "rand hex length mismatch"; exit 1; }
  b64=$($BIN rand -s 16 --b64)
  [ -n "$b64" ] || { echo "rand b64 empty"; exit 1; }
  b64url=$($BIN rand -s 16 --b64url)
  [ -n "$b64url" ] || { echo "rand b64url empty"; exit 1; }
  echo "  OK: rand generation"
}

function cmd_aes() {
  ensure_binary
  echo "--- Testing: aes ---"
  local key enc dec
  key=$($BIN rand -s 32)
  enc=$($BIN aes encrypt -k "$key" "super secret data")
  dec=$($BIN aes decrypt -k "$key" "$enc")
  assert_eq "super secret data" "$dec" "aes encrypt/decrypt"
  echo "  OK: aes encrypt/decrypt"
}

function cmd_rsa() {
  ensure_binary
  echo "--- Testing: rsa ---"
  local priv="$TMP_DIR/rsa_priv.pem"
  local pub="$TMP_DIR/rsa_pub.pem"

  $BIN rsa gen -b 1024 --priv-out "$priv" --pub-out "$pub" > /dev/null
  [ -f "$priv" ] && [ -f "$pub" ]

  local pub_extracted
  pub_extracted=$($BIN rsa pub -k "$priv")
  echo "$pub_extracted" | grep -q "BEGIN PUBLIC KEY"

  local key_id
  key_id=$($BIN rsa id -k "$pub")
  [ -n "$key_id" ]

  local sig
  sig=$($BIN rsa sign -k "$priv" "rsa message")
  $BIN rsa verify -k "$pub" --sig "$sig" "rsa message" > /dev/null

  local sig_b64
  sig_b64=$($BIN rsa sign -k "$priv" --b64 "rsa message")
  $BIN rsa verify -k "$pub" --b64 --sig "$sig_b64" "rsa message" > /dev/null

  local enc dec
  enc=$($BIN rsa encrypt -k "$pub" --label "rsalbl" "confidential rsa")
  dec=$($BIN rsa decrypt -k "$priv" --label "rsalbl" "$enc")
  assert_eq "confidential rsa" "$dec" "rsa encrypt/decrypt"

  echo "  OK: rsa gen, pub, id, sign, verify, encrypt, decrypt"
}

function cmd_25519() {
  ensure_binary
  echo "--- Testing: 25519 ---"
  local priv="$TMP_DIR/25519_priv.pem"
  local pub="$TMP_DIR/25519_pub.pem"

  $BIN 25519 gen --priv-out "$priv" --pub-out "$pub" > /dev/null
  [ -f "$priv" ] && [ -f "$pub" ]

  local pub_extracted
  pub_extracted=$($BIN 25519 pub -k "$priv")
  echo "$pub_extracted" | grep -q "BEGIN PUBLIC KEY"

  local key_id
  key_id=$($BIN 25519 id -k "$pub")
  [ -n "$key_id" ]

  local sig
  sig=$($BIN 25519 sign -k "$priv" "ed25519 message")
  $BIN 25519 verify -k "$pub" --sig "$sig" "ed25519 message" > /dev/null

  local sig_b64
  sig_b64=$($BIN 25519 sign -k "$priv" --b64 "ed25519 message")
  $BIN 25519 verify -k "$pub" --b64 --sig "$sig_b64" "ed25519 message" > /dev/null

  local enc dec
  enc=$($BIN 25519 encrypt -k "$pub" --label "25519lbl" "confidential 25519")
  dec=$($BIN 25519 decrypt -k "$priv" --label "25519lbl" "$enc")
  assert_eq "confidential 25519" "$dec" "25519 encrypt/decrypt"

  echo "  OK: 25519 gen, pub, id, sign, verify, encrypt, decrypt"
}

function cmd_salt() {
  ensure_binary
  echo "--- Testing: salt ---"
  local bundle
  bundle=$($BIN salt gen "my-password")
  $BIN salt verify -o "my-password" "$bundle" > /dev/null
  if $BIN salt verify -o "wrong-password" "$bundle" > /dev/null 2>&1; then
    echo "ERROR: salt verify should have failed for wrong password"
    exit 1
  fi
  echo "  OK: salt gen and verify"
}

function cmd_jwt() {
  ensure_binary
  echo "--- Testing: jwt ---"
  local priv="$TMP_DIR/jwt_rsa_priv.pem"
  local pub="$TMP_DIR/jwt_rsa_pub.pem"
  $BIN rsa gen -b 1024 --priv-out "$priv" --pub-out "$pub" > /dev/null

  local token
  token=$($BIN jwt seal -k "$priv" --iss "auth" --sub "user1" --aud "servicea" --exp "1h")
  [ -n "$token" ]

  local inspect_out
  inspect_out=$($BIN jwt inspect "$token")
  echo "$inspect_out" | grep -q "user1"

  $BIN jwt verify -k "$pub" --iss "auth" --sub "user1" --aud "servicea" "$token" > /dev/null
  echo "  OK: jwt seal, inspect, verify"
}

function cmd_kr_rsa() {
  ensure_binary
  echo "--- Testing: kr-rsa ---"
  local kr_file="$TMP_DIR/kr_rsa.json"
  $BIN kr-rsa init -f "$kr_file" > /dev/null
  $BIN kr-rsa gen -f "$kr_file" --owner "alice" -b 1024 > /dev/null
  $BIN kr-rsa gen -f "$kr_file" --owner "bob" -b 1024 > /dev/null

  local list_out
  list_out=$($BIN kr-rsa list -f "$kr_file")
  echo "$list_out" | grep -q "alice"
  echo "$list_out" | grep -q "bob"

  local exp_out
  exp_out=$($BIN kr-rsa export -f "$kr_file")
  echo "$exp_out" | grep -q "version"

  echo "  OK: kr-rsa init, gen, list, export"
}

function cmd_kr_aes() {
  ensure_binary
  echo "--- Testing: kr-aes ---"
  local kr_file="$TMP_DIR/kr_aes.json"
  $BIN kr-aes init -f "$kr_file" > /dev/null
  $BIN kr-aes gen -f "$kr_file" --owner "servicea" -s 32 > /dev/null
  $BIN kr-aes gen-pair -f "$kr_file" --src "servicea" --dst "serviceb" -s 32 > /dev/null

  local list_out
  list_out=$($BIN kr-aes list -f "$kr_file")
  echo "$list_out" | grep -q "servicea"
  echo "$list_out" | grep -q "serviceb"

  local exp_out
  exp_out=$($BIN kr-aes export -f "$kr_file")
  echo "$exp_out" | grep -q "servicea"

  echo "  OK: kr-aes init, gen, gen-pair, list, export"
}

function cmd_jwtmsg() {
  ensure_binary
  echo "--- Testing: jwtmsg ---"
  local kr_file="$TMP_DIR/kr_jwtmsg.json"
  $BIN kr-rsa init -f "$kr_file" > /dev/null
  $BIN kr-rsa gen -f "$kr_file" --owner "alice" -b 1024 > /dev/null
  $BIN kr-rsa gen -f "$kr_file" --owner "bob" -b 1024 > /dev/null

  # Raw encrypt / decrypt
  local enc dec
  enc=$($BIN jwtmsg encrypt --kr "$kr_file" --src "alice" --dst "bob" "party message")
  dec=$($BIN jwtmsg decrypt --kr "$kr_file" --src "alice" --dst "bob" "$enc")
  assert_eq "party message" "$dec" "jwtmsg raw encrypt/decrypt"

  # Envelope send / receive
  local token dec_env
  token=$($BIN jwtmsg send --kr "$kr_file" --src "alice" --dst "bob" --sub "order-99" "envelope payload")
  dec_env=$($BIN jwtmsg receive --kr "$kr_file" --src "alice" --dst "bob" "$token")
  echo "$dec_env" | grep -q "order-99"

  echo "  OK: jwtmsg encrypt, decrypt, send, receive"
}

function cmd_jwtstate() {
  ensure_binary
  echo "--- Testing: jwtstate ---"
  local kr_file="$TMP_DIR/kr_jwtstate.json"
  $BIN kr-aes init -f "$kr_file" > /dev/null
  $BIN kr-aes gen -f "$kr_file" --owner "sessions" -s 32 > /dev/null

  # Raw encrypt / decrypt
  local enc dec
  enc=$($BIN jwtstate encrypt --kr "$kr_file" --src "sessions" "state token 123")
  dec=$($BIN jwtstate decrypt --kr "$kr_file" --src "sessions" "$enc")
  assert_eq "state token 123" "$dec" "jwtstate raw encrypt/decrypt"

  # Envelope send / receive
  local token dec_env
  token=$($BIN jwtstate send --kr "$kr_file" --src "sessions" --sub "sess-user-1" "active-session-data")
  dec_env=$($BIN jwtstate receive --kr "$kr_file" --src "sessions" "$token")
  echo "$dec_env" | grep -q "sess-user-1"

  echo "  OK: jwtstate encrypt, decrypt, send, receive"
}

function cmd_all() {
  echo "=========================================="
  echo "Running all cryptz CLI command tests"
  echo "=========================================="
  cmd_version
  cmd_b64
  cmd_b32
  cmd_hash
  cmd_rand
  cmd_aes
  cmd_rsa
  cmd_25519
  cmd_salt
  cmd_jwt
  cmd_kr_rsa
  cmd_kr_aes
  cmd_jwtmsg
  cmd_jwtstate
  echo "=========================================="
  echo "All cryptz CLI tests passed successfully!"
  echo "=========================================="
}

cd "$(dirname "$0")"; _cmd="${1?"cmd is required"}"; shift; "cmd_${_cmd}" "$@"
