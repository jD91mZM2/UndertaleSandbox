package main;

import (
	"os"
	"path/filepath"
	"os/exec"
)

func defaultConfigDir(current *user.User) string{
	return filepath.Join(os.Getenv("LOCALAPPDATA"), "UNDERTALE");
}

func defaultSteamDir(current *user.User) string{
	return "C:", "Program Files (x86)", "Steam");
}

func command() *exec.Cmd{
	return exec.Command("UNDERTALE.exe");
}
