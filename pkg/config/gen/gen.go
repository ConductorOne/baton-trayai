package main

import (
	"github.com/conductorone/baton-sdk/pkg/config"
	cfg "github.com/conductorone/baton-trayai/pkg/config"
)

func main() {
	config.Generate("trayai", cfg.Config)
}
