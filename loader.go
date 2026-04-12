package template

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"strings"
)

// Loader locates and loads template source code by name.
//
// Implementations must validate the name with [ValidateName] (which adds
// backslash and NUL rejection on top of fs.ValidPath) and return
// ErrInvalidTemplateName for any path that fails the check. Unknown names
// must return ErrTemplateNotFound.
//
// The returned resolved name is used as the cache key and should be
// stable and unique within a loader (e.g., include a layer prefix
// when chained).
type Loader interface {
	Open(name string) (source string, resolved string, err error)
}

// ValidateName checks that name is safe to use as a template path.
// It rejects:
//   - anything fs.ValidPath rejects (empty element, "..", absolute, trailing /)
//   - backslash (Windows path separator; forces forward-slash discipline)
//   - NUL byte (path injection)
//
// Loader implementations must call this on every name they receive.
func ValidateName(name string) error {
	if !fs.ValidPath(name) {
		return fmt.Errorf("%w: %q", ErrInvalidTemplateName, name)
	}
	if strings.ContainsAny(name, "\\\x00") {
		return fmt.Errorf("%w: %q", ErrInvalidTemplateName, name)
	}
	return nil
}

// NewMemoryLoader returns a Loader that serves templates from an
// in-memory map. Intended for tests and small pre-registered sets.
func NewMemoryLoader(files map[string]string) Loader {
	copied := make(map[string]string, len(files))
	maps.Copy(copied, files)
	return &memoryLoader{files: copied}
}

type memoryLoader struct {
	files map[string]string
}

func (l *memoryLoader) Open(name string) (string, string, error) {
	if err := ValidateName(name); err != nil {
		return "", "", err
	}
	src, ok := l.files[name]
	if !ok {
		return "", "", fmt.Errorf("%w: %q", ErrTemplateNotFound, name)
	}
	return src, name, nil
}

// NewDirLoader returns a Loader that reads templates from the given
// local directory, sandboxed by [os.Root]. Symbolic links cannot escape
// the root: following any link whose target lies outside dir results
// in an error.
//
// This is the default, recommended way to load templates from disk.
//
// For development workflows that deliberately require symlink
// following (theme dev, monorepo sharing), use
// [NewFSLoader] with [os.DirFS] and accept responsibility for the
// relaxed sandbox.
func NewDirLoader(dir string) (Loader, error) {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, fmt.Errorf("dir loader: %w", err)
	}
	return &dirLoader{root: root, dir: dir}, nil
}

type dirLoader struct {
	root *os.Root
	dir  string
}

func (l *dirLoader) Open(name string) (string, string, error) {
	if err := ValidateName(name); err != nil {
		return "", "", err
	}
	f, err := l.root.Open(name)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", "", fmt.Errorf("%w: %q", ErrTemplateNotFound, name)
		}
		return "", "", fmt.Errorf("dir loader: %w", err)
	}
	defer func() { _ = f.Close() }()
	data, err := io.ReadAll(f)
	if err != nil {
		return "", "", fmt.Errorf("dir loader: %w", err)
	}
	return string(data), name, nil
}

// NewFSLoader wraps any [fs.FS] as a Loader. Intended for already-
// sandboxed filesystems such as [embed.FS], [testing/fstest.MapFS],
// and [archive/zip.Reader].
//
// Warning: if you pass a non-sandboxed fs.FS (for example [os.DirFS]
// pointing at a real directory) the library cannot prevent symbolic
// links from escaping. Prefer [NewDirLoader] for local directories
// unless you deliberately need this escape hatch.
func NewFSLoader(fsys fs.FS) Loader {
	return &fsLoader{fsys: fsys}
}

type fsLoader struct {
	fsys fs.FS
}

func (l *fsLoader) Open(name string) (string, string, error) {
	if err := ValidateName(name); err != nil {
		return "", "", err
	}
	data, err := fs.ReadFile(l.fsys, name)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", "", fmt.Errorf("%w: %q", ErrTemplateNotFound, name)
		}
		return "", "", fmt.Errorf("fs loader: %w", err)
	}
	return string(data), name, nil
}

// NewChainLoader returns a Loader that queries the given loaders in
// order and returns the first one that has the requested template.
//
// Chain loaders are typically used to implement override layers like
// user > theme > builtin. Each hit's resolved name is prefixed with
// the layer index so the same name in different layers produces
// distinct cache keys in an [Engine].
func NewChainLoader(loaders ...Loader) Loader {
	return &chainLoader{loaders: loaders}
}

type chainLoader struct {
	loaders []Loader
}

func (l *chainLoader) Open(name string) (string, string, error) {
	if err := ValidateName(name); err != nil {
		return "", "", err
	}
	for i, sub := range l.loaders {
		src, resolved, err := sub.Open(name)
		if err == nil {
			return src, fmt.Sprintf("layer%d:%s", i, resolved), nil
		}
		if errors.Is(err, ErrTemplateNotFound) {
			continue
		}
		return "", "", err
	}
	return "", "", fmt.Errorf("%w: %q", ErrTemplateNotFound, name)
}
