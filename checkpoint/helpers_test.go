// CRC: crc-Checkpoint.md | Test: test-Checkpoint.md
package checkpoint

import (
	"crypto/sha256"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Tests run the real fossil in temporary directories: the defects this package
// guards against were interactions between our command sequences and fossil's
// actual behaviour, which a fake fossil would only re-encode.

func testFossil(t *testing.T) string {
	t.Helper()
	fossil, err := Find()
	if err != nil {
		t.Skip("fossil is not installed")
	}
	return fossil
}

// newApp makes an app directory holding app.lua and returns its App.
func newApp(t *testing.T) App {
	t.Helper()
	a := App{Dir: filepath.Join(t.TempDir(), "demo"), Fossil: testFossil(t)}
	if err := os.Mkdir(a.Dir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, a.Dir, "app.lua", "-- v1\n")
	return a
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, dir, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// fossilIn runs fossil in dir and fails the test on error.
func fossilIn(t *testing.T, fossil, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command(fossil, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("fossil %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return string(out)
}

// catRev returns a file's content at a revision of the app's repository.
func catRev(t *testing.T, a App, name, rev string) string {
	t.Helper()
	return fossilIn(t, a.Fossil, a.Dir, "cat", name, "-r", rev)
}

func mustSave(t *testing.T, a App, msg string) {
	t.Helper()
	if _, err := a.Save(msg); err != nil {
		t.Fatalf("Save(%q): %v", msg, err)
	}
}

// fossilShim wraps the real fossil so a test can make it fail on demand:
// FAIL_ON=export or import fails that bundle step; import-noop makes bundle
// import exit 0 without importing.
const fossilShim = `#!/bin/sh
case "$FAIL_ON:$1 $2" in
    "export:bundle export"|"import:bundle import") echo "shim: injected $FAIL_ON failure" >&2; exit 1 ;;
    "import-noop:bundle import") exit 0 ;;
esac
exec "%s" "$@"
`

// withShim returns the app using a failing fossil, set to failOn.
func withShim(t *testing.T, a App, failOn string) App {
	t.Helper()
	shim := filepath.Join(t.TempDir(), "fossil")
	if err := os.WriteFile(shim, []byte(strings.Replace(fossilShim, "%s", a.Fossil, 1)), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("FAIL_ON", failOn)
	return App{Dir: a.Dir, Fossil: shim}
}

// snapshot hashes the repository and checkout database.
func snapshot(t *testing.T, a App) map[string][32]byte {
	t.Helper()
	sums := map[string][32]byte{}
	for _, p := range []string{a.repo(), a.checkout()} {
		data, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		sums[filepath.Base(p)] = sha256.Sum256(data)
	}
	return sums
}

func assertNoAside(t *testing.T, a App) {
	t.Helper()
	if m, _ := filepath.Glob(filepath.Join(a.Dir, ".#baseline.*")); len(m) > 0 {
		t.Errorf("aside directory left behind: %v", m)
	}
}
