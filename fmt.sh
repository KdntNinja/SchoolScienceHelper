for f in $(find . -type f -name '*.templ'); do
  templ fmt "$f"
done
