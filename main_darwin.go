package main;

import (
	"os/user"
	"path/filepath"
)

func defaultConfigDir(current *user.User) string{
	return filepath.Join(current.HomeDir, "Library", "Application Support", "com.tobyfox.undertale");
}

func defaultSteamDir(current *user.User) string{
	return filepath.Join(current.HomeDir, "Library", "Application Support", "Steam");
}
