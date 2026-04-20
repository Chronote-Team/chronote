#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(CDPATH="" cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../common.sh"

fail() {
    echo "FAIL: $*" >&2
    exit 1
}

assert_ok() {
    local description="$1"
    shift

    if ! "$@"; then
        fail "$description"
    fi
}

assert_eq() {
    local description="$1"
    local expected="$2"
    local actual="$3"

    if [[ "$expected" != "$actual" ]]; then
        fail "$description: expected '$expected', got '$actual'"
    fi
}

REPO_ROOT=$(get_repo_root)

assert_ok "conventional branch names should be accepted" check_feature_branch "refactor/all" "true"
assert_ok "numbered feature branches should still be accepted" check_feature_branch "001-user-auth" "true"

feature_dir=$(find_feature_dir_by_prefix "$REPO_ROOT" "refactor/all")
assert_eq "conventional branches should resolve to nested specs path" \
    "$REPO_ROOT/specs/refactor/all" \
    "$feature_dir"

feature_dir=$(find_feature_dir_by_prefix "$REPO_ROOT" "001-user-auth")
assert_eq "numbered branches should keep prefix-based lookup" \
    "$REPO_ROOT/specs/001-user-auth" \
    "$feature_dir"

echo "PASS"
