package checkpoint

import (
	"errors"
	"io"
	"os"
	"path/filepath"
)

// Status is one app's checkpoint state.
// CRC: crc-Checkpoint.md | R218
type Status struct {
	Count      int  // trunk checkpoints, not counting the baseline
	Downloaded bool // the app has original.fossil
	Changed    bool // the files differ from original.fossil's trunk
}

// Status reads an app's checkpoint state for a display to poll. It names each
// repository it reads, so fossil cannot walk up into an enclosing checkout, and it
// never opens, verifies or configures a checkout: it writes nothing.
// CRC: crc-Checkpoint.md | Seq: seq-checkpoint.md#5.4 | R218, R219, R220
func (a App) Status() (Status, error) {
	var st Status
	if a.hasRepo() {
		n, err := a.trunkCount("-R", a.repo())
		if err != nil {
			return st, err
		}
		st.Count = n
	}
	st.Downloaded = exists(filepath.Join(a.Dir, "original.fossil"))
	if st.Downloaded {
		changed, err := a.Changed()
		if err != nil {
			return st, err
		}
		st.Changed = changed
	}
	return st, nil
}

// Install marks an app as installed from a download: a baseline, then a copy of
// checkpoint.fossil to original.fossil. The copy goes to a dot-prefixed temporary
// (which the ignore-glob hides) and is renamed into place, so a failure leaves no
// partial original.fossil.
// CRC: crc-Checkpoint.md | Seq: seq-checkpoint.md#6.2 | R221
func (a App) Install() (err error) {
	if _, err := a.Baseline(); err != nil {
		return err
	}
	src, err := os.Open(a.repo())
	if err != nil {
		return err
	}
	defer src.Close()
	tmp, err := os.CreateTemp(a.Dir, ".#original.")
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			os.Remove(tmp.Name())
		}
	}()
	_, err = io.Copy(tmp, src)
	if cerr := tmp.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return err
	}
	if err = os.Rename(tmp.Name(), filepath.Join(a.Dir, "original.fossil")); err != nil {
		return errors.Join(errors.New("installing original.fossil"), err)
	}
	return nil
}
