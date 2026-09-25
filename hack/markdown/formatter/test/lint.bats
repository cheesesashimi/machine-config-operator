#!/usr/bin/env bats

repo_root="$(git rev-parse --show-toplevel)"

setup() {
  test_dir="$(mktemp -d "$repo_root/bats-lint.XXXXXX")"
  printf '%s\n' '# Markdown lint test' '' 'This document follows the default Markdown lint rules.' >"$test_dir/input.md"
}

teardown() {
  rm -rf "$test_dir"
}

@test "markdownlint accepts a valid Markdown document" {
  run env LINT_TARGET="$test_dir" "$repo_root/hack/markdown/markdownlint.sh"

  [ "$status" -eq 0 ]
}
