#!/usr/bin/env bats

repo_root="$(git rev-parse --show-toplevel)"

setup() {
  test_dir="$(mktemp -d "$repo_root/bats-format.XXXXXX")"
  test_file="$test_dir/input.md"
  document_file="${test_dir#"$repo_root/"}/input.md"
  printf '%s\n' \
    '---' \
    'title: “Keep” – …' \
    '---' \
    '' \
    'This is “prose” – with — ellipsis … and space.' \
    '' \
    '- one' \
    '- two' \
    '' \
    '```' \
    '%%{init: {"theme": "dark"}}%%' \
    'flowchart TD' \
    '```' \
    '' \
    '```text' \
    'graph TD' \
    'const quote = “keep” -- …;' \
    '```' \
    '' \
    '```' \
    'graph TD' \
    '```' >"$test_file"
}

teardown() {
  rm -rf "$test_dir"
}

@test "format-document applies the intended transformations" {
  run "$repo_root/hack/markdown/formatter/format-document.sh" "$document_file"
  [ "$status" -eq 0 ]

  run grep -Fq 'title: “Keep” – …' "$test_file"
  [ "$status" -eq 0 ]
  run grep -Fq 'This is "prose" - with -- ellipsis ... and space.' "$test_file"
  [ "$status" -eq 0 ]
  run grep -Fq -- '- one' "$test_file"
  [ "$status" -eq 0 ]
  run grep -Fq '```mermaid' "$test_file"
  [ "$status" -eq 0 ]
  run grep -Fq '%%{init: {"theme": "dark"}}%%' "$test_file"
  [ "$status" -eq 0 ]
  run grep -Fq '```text' "$test_file"
  [ "$status" -eq 0 ]
  run grep -Fq 'const quote = “keep” -- …;' "$test_file"
  [ "$status" -eq 0 ]
}
