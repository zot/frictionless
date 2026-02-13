//go:build windows

package mcp

import "syscall"

// detachedProcAttr returns SysProcAttr for a detached background process on Windows.
func detachedProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{CreationFlags: 0x00000008} // CREATE_NO_WINDOW
}
