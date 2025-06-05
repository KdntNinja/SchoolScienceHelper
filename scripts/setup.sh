clear
sudo dnf install golang -y
go install github.com/a-h/templ/cmd/templ@latest
curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64
chmod +x tailwindcss-linux-x64
sudo mv tailwindcss-linux-x64 /usr/local/bin/tailwindcss
go install github.com/air-verse/air@latest
go install github.com/axzilla/templui/cmd/templui@latest
