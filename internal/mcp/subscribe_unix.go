//go:build !windows

package mcp

import "syscall"

// detachedProcAttr returns SysProcAttr for a detached background process on Unix.
func detachedProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setsid: true}
}
