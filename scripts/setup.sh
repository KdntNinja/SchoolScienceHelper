#!/bin/bash
set -e

clear

# Check for sudo
if ! command -v sudo >/dev/null 2>&1; then
    echo "sudo is required. Please install sudo and re-run this script."
    exit 1
fi

# Install dependencies
if command -v apt >/dev/null 2>&1; then
    sudo apt update
    sudo apt install -y golang curl git
elif command -v dnf >/dev/null 2>&1; then
    sudo dnf install -y golang curl git
else
    echo "Neither apt nor dnf found. Please install Go, curl, and git manually."
    exit 1
fi

# Install Go tools
go install github.com/a-h/templ/cmd/templ@latest
go install github.com/air-verse/air@latest
go install github.com/axzilla/templui/cmd/templui@latest

# Install TailwindCSS
curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64
chmod +x tailwindcss-linux-x64
sudo mv tailwindcss-linux-x64 /usr/local/bin/tailwindcss

echo "Setup complete."
