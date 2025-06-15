#!/usr/bin/env bash

clear
find . -type f -name '*.templ' -print0 | while IFS= read -r -d '' f; do
  templ fmt "$f"
done

# Find all directories containing Go files and run go fmt in each
find . -type f -name '*.go' -exec dirname {} \; | sort -u | while read -r dir; do
  (cd "$dir" && go fmt)
done
