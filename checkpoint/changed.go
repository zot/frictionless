package checkpoint

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// fossilFiles never count as app files.
var fossilFiles = map[string]bool{
	"checkpoint.fossil": true, "original.fossil": true, ".fslckout": true, "_FOSSIL_": true,
}

// ignored reports whether a path relative to the app directory is fossil's own
// or matches the checkpoint ignore-glob.
// CRC: crc-Checkpoint.md | R216
func ignored(rel string) bool {
	base := path.Base(rel)
	if fossilFiles[base] {
		return true
	}
	for _, g := range strings.Split(ignoreGlob, ",") {
		if ok, _ := path.Match(g, base); ok {
			return true
		}
		if ok, _ := path.Match(g, rel); ok {
			return true
		}
	}
	return false
}

// Changed reports whether the app's files differ from the tip of trunk in
// original.fossil: a file changed, added or removed. An app without
// original.fossil was not downloaded and is never changed.
// CRC: crc-Checkpoint.md | Seq: seq-checkpoint.md#4.1 | R215, R216, R217
func (a App) Changed() (bool, error) {
	orig := filepath.Join(a.Dir, "original.fossil")
	if !exists(orig) {
		return false, nil // Seq: seq-checkpoint.md#4.1
	}
	installed, err := a.trunkFiles(orig)
	if err != nil {
		return false, err
	}
	// Seq: seq-checkpoint.md#4.4
	changed := false
	err = filepath.WalkDir(a.Dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || changed {
			return err
		}
		rel, _ := filepath.Rel(a.Dir, p)
		rel = filepath.ToSlash(rel)
		if rel == "." {
			return nil
		}
		if ignored(rel) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}
		want, ok := installed[rel]
		if !ok {
			changed = true // added
			return nil
		}
		delete(installed, rel)
		got, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		changed = !bytes.Equal(got, want)
		return nil
	})
	if err != nil {
		return false, err
	}
	// Seq: seq-checkpoint.md#4.5 — anything left in installed was removed
	return changed || len(installed) > 0, nil
}

// trunkFiles exports the tip of trunk in repo as one tarball and returns its
// files by path, with the tarball's top directory stripped.
// Seq: seq-checkpoint.md#4.2, seq-checkpoint.md#4.3
func (a App) trunkFiles(repo string) (map[string][]byte, error) {
	tmp, err := os.CreateTemp("", "checkpoint-original-*.tar.gz")
	if err != nil {
		return nil, err
	}
	tmp.Close()
	defer os.Remove(tmp.Name())
	if _, err := a.step("tarball", "trunk", tmp.Name(), "-R", repo); err != nil {
		return nil, err
	}
	f, err := os.Open(tmp.Name())
	if err != nil {
		return nil, err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	files := map[string][]byte{}
	tr := tar.NewReader(gz)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			return files, nil
		}
		if err != nil {
			return nil, err
		}
		if h.Typeflag != tar.TypeReg {
			continue
		}
		_, rel, ok := strings.Cut(h.Name, "/")
		if !ok || ignored(rel) {
			continue
		}
		data, err := io.ReadAll(tr)
		if err != nil {
			return nil, err
		}
		files[rel] = data
	}
}
