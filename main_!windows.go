// +build !windows

package main

import (
	"os/exec"
)

func command() *exec.Cmd {
	return exec.Command("./runner")
}
