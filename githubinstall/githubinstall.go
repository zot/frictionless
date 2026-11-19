// Package githubinstall investigates and installs an app from a GitHub tree URL
// with Go's HTTP client and archive/zip; it runs no other program
// (specs/github-install.md).
// CRC: crc-GitHubInstall.md | R246
package githubinstall

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// The endpoints. They are variables so tests can serve them locally. No
// credentials are sent: these are GitHub's public endpoints.
// CRC: crc-GitHubInstall.md | R247
var (
	APIBase     = "https://api.github.com"
	RawBase     = "https://raw.githubusercontent.com"
	ArchiveBase = "https://github.com"
)

// MaxBytes caps both the downloaded archive and the total of the files
// extracted from it. A variable so tests can lower it.
// CRC: crc-GitHubInstall.md | R243
var MaxBytes int64 = 100 << 20

// Source is one directory in one branch of a GitHub repository.
type Source struct {
	User, Repo, Branch, Path string
	Name                     string // the last segment of Path: the app's name
}

// Entry is one item of a directory listing.
type Entry struct {
	Name string `json:"name"`
	Type string `json:"type"` // "file" or "dir"
}

// Investigation is what a review needs: both listings and the text of every
// reviewable file, keyed by its path relative to the app directory.
type Investigation struct {
	Files, Viewdefs []Entry
	Contents        map[string]string
}

// ParseURL reads https://github.com/{user}/{repo}/tree/{branch}/{path}; any
// other shape is an error.
// CRC: crc-GitHubInstall.md | R236
func ParseURL(raw string) (Source, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Host != "github.com" {
		return Source{}, fmt.Errorf("not a GitHub tree URL: %s", raw)
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 5 || parts[2] != "tree" {
		return Source{}, fmt.Errorf("not a GitHub tree URL (want https://github.com/user/repo/tree/branch/path): %s", raw)
	}
	for _, p := range parts {
		if p == "" || p == "." || p == ".." {
			return Source{}, fmt.Errorf("not a GitHub tree URL: %s", raw)
		}
	}
	p := strings.Join(parts[4:], "/")
	return Source{User: parts[0], Repo: parts[1], Branch: parts[3], Path: p, Name: parts[len(parts)-1]}, nil
}

// get fetches url, refusing a non-200 status and a body larger than MaxBytes.
func get(ctx context.Context, rawURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: %s", rawURL, resp.Status)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, MaxBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > MaxBytes {
		return nil, fmt.Errorf("%s: larger than %d bytes", rawURL, MaxBytes)
	}
	return body, nil
}

// list fetches a directory listing from the contents API.
// CRC: crc-GitHubInstall.md | R237
func list(ctx context.Context, src Source, dir string) ([]Entry, error) {
	u := fmt.Sprintf("%s/repos/%s/%s/contents/%s?ref=%s", APIBase, src.User, src.Repo, dir, url.QueryEscape(src.Branch))
	body, err := get(ctx, u)
	if err != nil {
		return nil, err
	}
	var entries []Entry
	if err := json.Unmarshal(body, &entries); err != nil {
		return nil, fmt.Errorf("listing %s: %w", dir, err)
	}
	return entries, nil
}

// fetch returns a file's text, or false when it cannot be fetched.
func fetch(ctx context.Context, src Source, rel string) (string, bool) {
	u := fmt.Sprintf("%s/%s/%s/%s/%s/%s", RawBase, src.User, src.Repo, src.Branch, src.Path, rel)
	body, err := get(ctx, u)
	if err != nil {
		return "", false
	}
	return string(body), true
}

// Investigate lists the app directory and its viewdefs/, and fetches every file
// the review can show. A file that cannot be fetched is left out; an
// unfetchable app-directory listing is the error.
// CRC: crc-GitHubInstall.md | Seq: seq-github-install.md#1.2 | R237, R238, R239
func Investigate(ctx context.Context, src Source) (Investigation, error) {
	files, err := list(ctx, src, src.Path)
	if err != nil {
		return Investigation{}, err
	}
	viewdefs, err := list(ctx, src, src.Path+"/viewdefs")
	if err != nil {
		viewdefs = nil // no viewdefs/ is the review's problem, not a fetch error
	}
	inv := Investigation{Files: files, Viewdefs: viewdefs, Contents: map[string]string{}}
	var want []string
	for _, f := range files {
		if f.Type == "file" && (f.Name == "requirements.md" || f.Name == "design.md" || strings.HasSuffix(f.Name, ".lua")) {
			want = append(want, f.Name)
		}
	}
	for _, f := range viewdefs {
		if f.Type == "file" && strings.HasSuffix(f.Name, ".html") {
			want = append(want, "viewdefs/"+f.Name)
		}
	}
	for _, rel := range want {
		if text, ok := fetch(ctx, src, rel); ok {
			inv.Contents[rel] = text
		}
	}
	return inv, nil
}

// AsLua is the plain-value form delivered to Lua: {files, viewdefs, contents},
// each entry {name, type}.
// CRC: crc-GitHubInstall.md | R239
func (inv Investigation) AsLua() map[string]any {
	entries := func(es []Entry) []any {
		out := make([]any, len(es))
		for i, e := range es {
			out[i] = map[string]any{"name": e.Name, "type": e.Type}
		}
		return out
	}
	contents := make(map[string]any, len(inv.Contents))
	for k, v := range inv.Contents {
		contents[k] = v
	}
	return map[string]any{"files": entries(inv.Files), "viewdefs": entries(inv.Viewdefs), "contents": contents}
}

// Install downloads the branch archive and extracts the app's directory into
// appsDir/name. It never overwrites an existing app, and apps/name is either the
// complete app or absent.
// CRC: crc-GitHubInstall.md | Seq: seq-github-install.md#2.2 | R240, R241, R243, R244, R248
func Install(ctx context.Context, src Source, appsDir, name string) error {
	if name == "" || name == "." || strings.ContainsAny(name, `/\`) || strings.Contains(name, "..") {
		return fmt.Errorf("invalid app name %q", name)
	}
	dest := filepath.Join(appsDir, name)
	if _, err := os.Lstat(dest); err == nil {
		return fmt.Errorf("app %q already exists in %s", name, appsDir)
	}
	archive, err := get(ctx, fmt.Sprintf("%s/%s/%s/archive/refs/heads/%s.zip", ArchiveBase, src.User, src.Repo, src.Branch))
	if err != nil {
		return err
	}
	tmp, err := os.MkdirTemp(appsDir, ".#install.")
	if err != nil {
		return err
	}
	if err := os.Chmod(tmp, 0o755); err != nil { // MkdirTemp makes it 0700; an app directory is 0755
		os.RemoveAll(tmp)
		return err
	}
	if err := extract(archive, src.Path, tmp); err != nil {
		os.RemoveAll(tmp)
		return err
	}
	if err := os.Rename(tmp, dest); err != nil {
		os.RemoveAll(tmp)
		return err
	}
	return nil
}

// extract writes the archive's entries under {top}/{appPath}/ into dir, with
// that prefix removed. An entry that would land outside dir, or a total over
// MaxBytes, fails the whole extraction.
// CRC: crc-GitHubInstall.md | R241, R242, R243
func extract(archive []byte, appPath, dir string) error {
	zr, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		return fmt.Errorf("reading the archive: %w", err)
	}
	var total int64
	found := false
	for _, f := range zr.File {
		top, rest, ok := strings.Cut(f.Name, "/")
		if !ok || top == "" {
			continue
		}
		rel, ok := strings.CutPrefix(rest, appPath+"/")
		if !ok || rel == "" {
			continue
		}
		found = true
		clean := path.Clean(rel)
		if path.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, "../") {
			return fmt.Errorf("archive entry %q escapes the app directory", f.Name)
		}
		target := filepath.Join(dir, filepath.FromSlash(clean))
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		n, err := writeEntry(f, target, MaxBytes-total)
		if err != nil {
			return err
		}
		total += n
	}
	if !found {
		return errors.New("the archive holds no files under " + appPath)
	}
	return nil
}

// writeEntry copies one entry to target, refusing more than limit bytes.
func writeEntry(f *zip.File, target string, limit int64) (int64, error) {
	rc, err := f.Open()
	if err != nil {
		return 0, err
	}
	defer rc.Close()
	out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return 0, err
	}
	n, err := io.Copy(out, io.LimitReader(rc, limit+1))
	if cerr := out.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return n, err
	}
	if n > limit {
		return n, fmt.Errorf("the app's files total more than %d bytes", MaxBytes)
	}
	return n, nil
}
