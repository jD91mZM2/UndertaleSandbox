package main

import (
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
)

func defaultConfigDir(current *user.User) string {
	return filepath.Join(os.Getenv("LOCALAPPDATA"), "UNDERTALE")
}

func defaultSteamDir(current *user.User) string {
	return filepath.Join("C:", "Program Files (x86)", "Steam")
}

func command() *exec.Cmd {
	return exec.Command("UNDERTALE.exe")
}
