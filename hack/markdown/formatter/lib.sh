#!/usr/bin/env bash

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

parse_document_reference() {
  # Keep URL and ID normalization identical for upload and download operations.
  local document_reference="$1"
  local document_id="$document_reference"

  if [[ "$document_reference" == */d/* ]]; then
    document_id="${document_reference#*/d/}"
    document_id="${document_id%%/*}"
    document_id="${document_id%%\?*}"
    document_id="${document_id%%\#*}"
  fi
  if [[ ! "$document_id" =~ ^[A-Za-z0-9_-]+$ ]]; then
    printf 'invalid Google document URL or ID: %s\n' "$document_reference" >&2
    return 2
  fi

  printf '%s\n' "$document_id"
}

validate_markdown_path() {
  # Reject traversal before callers construct paths under the workspace.
  local markdown_file="$1"
  local location="${2:-workspace}"

  if [[ "$markdown_file" = /* || "$markdown_file" == .. || "$markdown_file" == ../* || "$markdown_file" == */../* || "$markdown_file" != *.md ]]; then
    printf 'Markdown file must be a relative .md path under the %s: %s\n' "$location" "$markdown_file" >&2
    return 2
  fi
}

validate_frontmatter() {
  # Keep YAML validation alongside the shared shell helpers used by the upload path.
  node "$script_dir/config/validate-frontmatter.mjs" "$1"
}

strip_frontmatter() {
  local repo_file="$1"
  local frontmatter_file="$2"
  local stripped_file="$3"

  # Preserve the whole file when a leading frontmatter block is not closed.
  awk -v frontmatter_file="$frontmatter_file" -v stripped_file="$stripped_file" '
    { lines[NR] = $0 }
    NR == 1 && $0 == "---" { in_frontmatter = 1; next }
    in_frontmatter && $0 == "---" { closing_delimiter = NR; in_frontmatter = 0 }
    END {
      if (closing_delimiter) {
        for (line = 1; line <= closing_delimiter; line++) {
          print lines[line] > frontmatter_file
        }
        for (line = closing_delimiter + 1; line <= NR; line++) {
          print lines[line] > stripped_file
        }
      } else {
        for (line = 1; line <= NR; line++) {
          print lines[line] > stripped_file
        }
      }
    }
  ' "$repo_file"
}

extract_frontmatter() {
  local repo_file="$1"

  # Emit frontmatter only when its closing delimiter is present.
  awk '
    { lines[NR] = $0 }
    NR == 1 && $0 == "---" { in_frontmatter = 1; next }
    in_frontmatter && $0 == "---" { closing_delimiter = NR; in_frontmatter = 0 }
    END {
      if (closing_delimiter) {
        for (line = 1; line <= closing_delimiter; line++) {
          print lines[line]
        }
      }
    }
  ' "$repo_file"
}
