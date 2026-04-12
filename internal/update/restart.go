package update

import (
	"os"
	"syscall"
)

// RestartSelf re-executes the current binary using syscall.Exec.
func RestartSelf(binPath string) error {
	return syscall.Exec(binPath, os.Args, os.Environ())
}
