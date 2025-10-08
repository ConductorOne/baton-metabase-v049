package main

import (
	cfg "github.com/conductorone/baton-metabase-v049/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/config"
)

func main() {
	config.Generate("metabase-v049", cfg.Config)
}
