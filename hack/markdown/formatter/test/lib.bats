#!/usr/bin/env bats

repo_root="$(git rev-parse --show-toplevel)"
load "$repo_root/hack/markdown/formatter/lib.sh"

setup() {
  test_dir="$(mktemp -d)"
}

teardown() {
  rm -rf "$test_dir"
}

@test "parses a Google document ID and URL" {
  [ "$(parse_document_reference 'document-id')" = 'document-id' ]
  [ "$(parse_document_reference 'https://docs.google.com/document/d/document-id/edit')" = 'document-id' ]
}

@test "rejects an invalid Google document reference" {
  run parse_document_reference 'not a document ID'

  [ "$status" -eq 2 ]
}

@test "validates relative Markdown paths" {
  validate_markdown_path 'docs/example.md'
}

@test "rejects unsafe or non-Markdown paths" {
  run validate_markdown_path '../example.md'
  [ "$status" -eq 2 ]

  run validate_markdown_path 'docs/example.txt'
  [ "$status" -eq 2 ]
}

@test "strips closed frontmatter and preserves it separately" {
  input_file="$test_dir/input.md"
  frontmatter_file="$test_dir/frontmatter.md"
  stripped_file="$test_dir/stripped.md"
  printf '%s\n' '---' 'title: Example' '---' '' '# Body' >"$input_file"

  strip_frontmatter "$input_file" "$frontmatter_file" "$stripped_file"

  run cat "$frontmatter_file"
  [ "$output" = $'---\ntitle: Example\n---' ]
  run cat "$stripped_file"
  [ "$output" = $'\n# Body' ]
}

@test "preserves an unclosed block as document content" {
  input_file="$test_dir/input.md"
  frontmatter_file="$test_dir/frontmatter.md"
  stripped_file="$test_dir/stripped.md"
  printf '%s\n' '---' 'title: Example' '# Body' >"$input_file"

  strip_frontmatter "$input_file" "$frontmatter_file" "$stripped_file"

  [ ! -s "$frontmatter_file" ]
  run cat "$stripped_file"
  [ "$output" = $'---\ntitle: Example\n# Body' ]
}

@test "extracts only a closed frontmatter block" {
  input_file="$test_dir/input.md"
  printf '%s\n' '---' 'title: Example' '---' '# Body' >"$input_file"

  run extract_frontmatter "$input_file"

  [ "$output" = $'---\ntitle: Example\n---' ]
}

@test "does not extract an unclosed frontmatter block" {
  input_file="$test_dir/input.md"
  printf '%s\n' '---' 'title: Example' '# Body' >"$input_file"

  run extract_frontmatter "$input_file"

  [ -z "$output" ]
}

@test "validates frontmatter YAML" {
  valid_file="$test_dir/valid.md"
  invalid_file="$test_dir/invalid.md"
  printf '%s\n' '---' 'title: Example' '---' >"$valid_file"
  printf '%s\n' '---' 'title: [invalid' '---' >"$invalid_file"

  validate_frontmatter "$valid_file"
  run validate_frontmatter "$invalid_file"
  [ "$status" -ne 0 ]
}
