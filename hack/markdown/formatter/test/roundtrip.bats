#!/usr/bin/env bats

repo_root="$(git rev-parse --show-toplevel)"
load "$repo_root/hack/markdown/formatter/lib.sh"

document_url="${GWS_TEST_DOCUMENT_URL:-}"

setup_file() {
  if [[ -n "${ROUNDTRIP_SKIP_REASON:-}" ]]; then
    skip "$ROUNDTRIP_SKIP_REASON"
  fi
  if [[ -z "${GWS_TEST_DOCUMENT_URL:-}" ]]; then
    skip 'DOCUMENT_URL is not set'
  fi
  if [[ ! -d "${GWS_CONFIG_DIR:-/tmp/.config/gws}" ]]; then
    skip 'Google Workspace credentials are unavailable'
  fi
}

setup() {
  test_dir="$(mktemp -d "$repo_root/bats-roundtrip.XXXXXX")"
  document_file="${test_dir#"$repo_root/"}/input.md"
}

teardown() {
  rm -rf "$test_dir"
}

roundtrip() {
  local with_frontmatter="$1"
  local expected_file="$BATS_TEST_TMPDIR/expected.md"

  if [[ "$with_frontmatter" == true ]]; then
    printf '%s\n' '---' 'something: hello' '---' '' '# Formatter integration test' '' '```' 'graph TD' '    A[Start] --> B[Done]' '```' >"$repo_root/$document_file"
  else
    printf '%s\n' '# Formatter integration test' '' '```' 'graph TD' '    A[Start] --> B[Done]' '```' >"$repo_root/$document_file"
  fi
  run "$repo_root/hack/markdown/formatter/format-document.sh" "$document_file"
  [ "$status" -eq 0 ]
  cp "$repo_root/$document_file" "$expected_file"

  run "$repo_root/hack/markdown/formatter/push-to-google-drive.sh" "$document_url" "$document_file"
  [ "$status" -eq 0 ]

  run "$repo_root/hack/markdown/formatter/pull-from-google-drive.sh" "$document_url" "$document_file"
  [ "$status" -eq 0 ]
  expected_hash="$(sha256sum "$expected_file" | awk '{print $1}')"
  actual_hash="$(sha256sum "$repo_root/$document_file" | awk '{print $1}')"
  [ "$expected_hash" = "$actual_hash" ]
}

@test "roundtrips Markdown with frontmatter through Google Drive" {
  roundtrip true
}

@test "roundtrips Markdown without frontmatter through Google Drive" {
  roundtrip false
}
