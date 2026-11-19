//go:build windows

package applink

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"unicode/utf16"

	"golang.org/x/sys/windows"
)

// makeLink creates the activation link as a directory junction: a mount-point
// reparse point written with FSCTL_SET_REPARSE_POINT. A junction needs no
// SeCreateSymbolicLinkPrivilege and no Developer Mode, and this runs no other
// program. Its target is necessarily absolute, which is why Repair exists.
// Lifted from ark's dirlink_windows.go.
// CRC: crc-AppLink.md | Seq: seq-applink.md#1.1.5 | R227
func makeLink(linkPath, appDir, app string) error {
	target, err := filepath.Abs(appDir)
	if err != nil {
		return err
	}
	// A junction is a reparse point on an (empty) directory.
	if err := os.Mkdir(linkPath, 0755); err != nil && !os.IsExist(err) {
		return err
	}
	linkPtr, err := windows.UTF16PtrFromString(linkPath)
	if err != nil {
		return err
	}
	handle, err := windows.CreateFile(
		linkPtr,
		windows.GENERIC_WRITE,
		0,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_FLAG_OPEN_REPARSE_POINT|windows.FILE_FLAG_BACKUP_SEMANTICS,
		0,
	)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(handle)

	// The substitute name is the absolute target as an NT path (\??\C:\...);
	// the print name is the plain absolute path shown to the user.
	target = filepath.Clean(target)
	subst := utf16.Encode([]rune(`\??\` + target))
	print := utf16.Encode([]rune(target))

	// PathBuffer = substitute-name NUL print-name NUL (all UTF-16LE).
	pathBuf := make([]uint16, 0, len(subst)+len(print)+2)
	pathBuf = append(pathBuf, subst...)
	pathBuf = append(pathBuf, 0)
	pathBuf = append(pathBuf, print...)
	pathBuf = append(pathBuf, 0)

	substLen := len(subst) * 2
	printOffset := (len(subst) + 1) * 2
	printLen := len(print) * 2

	// MountPointReparseBuffer header (four USHORT) + PathBuffer.
	reparseDataLen := 8 + len(pathBuf)*2
	buf := make([]byte, 8+reparseDataLen)
	binary.LittleEndian.PutUint32(buf[0:], windows.IO_REPARSE_TAG_MOUNT_POINT)
	binary.LittleEndian.PutUint16(buf[4:], uint16(reparseDataLen))
	binary.LittleEndian.PutUint16(buf[6:], 0) // Reserved
	binary.LittleEndian.PutUint16(buf[8:], 0) // substitute-name offset
	binary.LittleEndian.PutUint16(buf[10:], uint16(substLen))
	binary.LittleEndian.PutUint16(buf[12:], uint16(printOffset))
	binary.LittleEndian.PutUint16(buf[14:], uint16(printLen))
	for i, w := range pathBuf {
		binary.LittleEndian.PutUint16(buf[16+i*2:], w)
	}

	var bytesReturned uint32
	return windows.DeviceIoControl(
		handle,
		windows.FSCTL_SET_REPARSE_POINT,
		&buf[0],
		uint32(len(buf)),
		nil,
		0,
		&bytesReturned,
		nil,
	)
}
