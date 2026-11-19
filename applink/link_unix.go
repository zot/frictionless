//go:build !windows

package applink

import (
	"os"
	"path/filepath"
)

// makeLink creates the activation link as a symbolic link with the relative
// target ../apps/{app}, so a moved base directory keeps it resolving.
// CRC: crc-AppLink.md | Seq: seq-applink.md#1.1.5 | R226
func makeLink(linkPath, appDir, app string) error {
	return os.Symlink(filepath.Join("..", appsDir, app), linkPath)
}
