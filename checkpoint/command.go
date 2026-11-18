package checkpoint

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

const usage = "usage: checkpoint save|list|rollback|diff|clear|baseline|count|update|local APP [ARG]"

// Command runs one checkpoint command, `CMD APP [ARG]`, for the app appsDir/APP,
// printing what the command reports to out. onChange, when non-nil, is called
// after a command that changed checkpoints.
// CRC: crc-Checkpoint.md | Seq: seq-checkpoint.md#1.1 | R79, R204, R208, R209
func Command(args []string, appsDir string, out io.Writer, onChange func()) error {
	if len(args) < 2 {
		return errors.New(usage)
	}
	cmd, name, arg := args[0], args[1], ""
	if len(args) > 2 {
		arg = args[2]
	}
	fossil, err := Find()
	if err != nil {
		return err
	}
	a := App{Dir: filepath.Join(appsDir, name), Fossil: fossil}
	if info, err := os.Stat(a.Dir); err != nil || !info.IsDir() {
		return fmt.Errorf("app not found: %s", name)
	}
	changed, err := a.command(cmd, arg, out)
	if err != nil {
		return err
	}
	if changed && onChange != nil {
		onChange()
	}
	return nil
}

// command runs one verb and reports whether it changed checkpoints.
func (a App) command(cmd, arg string, out io.Writer) (changed bool, err error) {
	n := 0
	if (cmd == "rollback" || cmd == "diff") && arg != "" {
		if n, err = strconv.Atoi(arg); err != nil {
			return false, fmt.Errorf("checkpoint %s not found", arg)
		}
	}
	switch cmd {
	case "save":
		created := !a.hasRepo()
		saved, err := a.Save(arg)
		switch {
		case err != nil:
			return false, err
		case created:
			fmt.Fprintf(out, "Created checkpoint: %s\n", orDefault(arg, "initial state"))
		case saved:
			fmt.Fprintf(out, "Saved checkpoint: %s\n", orDefault(arg, "checkpoint"))
		default:
			fmt.Fprintln(out, "No changes to checkpoint")
		}
		return saved, nil
	case "list":
		lines, err := a.List()
		if err != nil {
			return false, err
		}
		if !a.hasRepo() {
			fmt.Fprintf(out, "No checkpoints for %s\n", a.name())
		}
		for _, l := range lines {
			fmt.Fprintln(out, l)
		}
		return false, nil
	case "rollback":
		if err := a.Rollback(n); err != nil {
			return false, err
		}
		fmt.Fprintln(out, "Rolled back to checkpoint")
		return true, nil
	case "diff":
		if n == 0 {
			n = 1
		}
		d, err := a.Diff(n)
		if err != nil {
			return false, err
		}
		fmt.Fprint(out, d)
		return false, nil
	case "clear", "baseline":
		restored, err := a.Baseline()
		if err != nil {
			return false, err
		}
		if len(restored) > 0 {
			fmt.Fprintf(out, "Baseline set for %s (preserved branches restored)\n", a.name())
		} else {
			fmt.Fprintf(out, "Baseline set for %s\n", a.name())
		}
		return true, nil
	case "count":
		c, err := a.Count()
		if err != nil {
			return false, err
		}
		fmt.Fprintln(out, c)
		return false, nil
	case "update":
		if err := a.Update(arg); err != nil {
			return false, err
		}
		fmt.Fprintf(out, "Update checkpoint: %s\n", orDefault(arg, "update"))
		return true, nil
	case "local":
		if err := a.Local(arg); err != nil {
			return false, err
		}
		fmt.Fprintf(out, "Local checkpoint: %s\n", orDefault(arg, "local"))
		return true, nil
	}
	return false, fmt.Errorf("unknown checkpoint command: %s\n%s", cmd, usage)
}

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
