package assets

import "embed"

//go:embed css/* js/components/*
var Assets embed.FS
