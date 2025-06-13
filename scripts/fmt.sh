#!/usr/bin/env bash

clear
find . -type f -name '*.templ' -print0 | while IFS= read -r -d '' f; do
  templ fmt "$f"
done
