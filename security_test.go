package template

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// Phase L cycle 1: NewDirLoader happy path reads a file within root.
func TestDirLoader_HappyPath(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}

	loader, err := NewDirLoader(dir)
	if err != nil {
		t.Fatalf("NewDirLoader err = %v", err)
	}
	src, resolved, err := loader.Open("a.txt")
	if err != nil {
		t.Fatalf("Open err = %v", err)
	}
	if src != "hello" {
		t.Errorf("src = %q", src)
	}
	if resolved != "a.txt" {
		t.Errorf("resolved = %q", resolved)
	}
}

// Phase L cycle 2: T1 — relative escape is rejected by ValidateName.
func TestDirLoader_RelativeEscape_Rejected(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	loader, _ := NewDirLoader(dir)
	_, _, err := loader.Open("../a.txt")
	if !errors.Is(err, ErrInvalidTemplateName) {
		t.Errorf("err = %v, want ErrInvalidTemplateName", err)
	}
}

// Phase L cycle 3: T2 — absolute path is rejected.
func TestDirLoader_AbsolutePath_Rejected(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	loader, _ := NewDirLoader(dir)
	_, _, err := loader.Open("/etc/passwd")
	if !errors.Is(err, ErrInvalidTemplateName) {
		t.Errorf("err = %v, want ErrInvalidTemplateName", err)
	}
}

// Phase L cycle 4: T3 — symlink escape is blocked by os.Root.
func TestDirLoader_SymlinkEscape_Rejected(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("symlink semantics differ on Windows")
	}

	outside := t.TempDir()
	secret := filepath.Join(outside, "secret.txt")
	if err := os.WriteFile(secret, []byte("TOP SECRET"), 0o600); err != nil {
		t.Fatal(err)
	}

	root := t.TempDir()
	link := filepath.Join(root, "leak.txt")
	if err := os.Symlink(secret, link); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	loader, err := NewDirLoader(root)
	if err != nil {
		t.Fatalf("NewDirLoader: %v", err)
	}
	src, _, err := loader.Open("leak.txt")
	if err == nil {
		t.Errorf("expected error, got src = %q", src)
	}
}

// Phase L cycle 5: T3b — nested symlink within a subdirectory is blocked.
func TestDirLoader_NestedSymlinkEscape_Rejected(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("symlink semantics differ on Windows")
	}

	outside := t.TempDir()
	secret := filepath.Join(outside, "passwd")
	_ = os.WriteFile(secret, []byte("root:x:0:"), 0o600)

	root := t.TempDir()
	sub := filepath.Join(root, "sub")
	_ = os.Mkdir(sub, 0o750)
	link := filepath.Join(sub, "leak.txt")
	if err := os.Symlink(secret, link); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	loader, _ := NewDirLoader(root)
	_, _, err := loader.Open("sub/leak.txt")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// Phase L cycle 6: T4 — NUL injection is rejected.
func TestDirLoader_NULInjection_Rejected(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	loader, _ := NewDirLoader(dir)
	_, _, err := loader.Open("a\x00b.txt")
	if !errors.Is(err, ErrInvalidTemplateName) {
		t.Errorf("err = %v, want ErrInvalidTemplateName", err)
	}
}

// Phase L cycle 7: T4b — newline in name is rejected.
func TestDirLoader_NewlineInjection_Rejected(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	loader, _ := NewDirLoader(dir)
	// fs.ValidPath rejects newline? Actually it doesn't; ValidateName
	// rejects NUL and backslash only. Newline is technically allowed by
	// fs.ValidPath. We document this: newlines don't escape the sandbox,
	// they just produce an impossible filename.
	_, _, err := loader.Open("a\nb.txt")
	// Any error is fine — either ValidateName or file-not-found.
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// Phase L cycle 8: T5 — Windows drive letter is rejected.
func TestDirLoader_WindowsDrive_Rejected(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	loader, _ := NewDirLoader(dir)
	_, _, err := loader.Open("C:\\Windows\\System32\\config\\SAM")
	if !errors.Is(err, ErrInvalidTemplateName) {
		t.Errorf("err = %v, want ErrInvalidTemplateName", err)
	}
}

// Phase L cycle 9: T5b — UNC path is rejected.
func TestDirLoader_UNCPath_Rejected(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	loader, _ := NewDirLoader(dir)
	_, _, err := loader.Open("\\\\server\\share\\x")
	if !errors.Is(err, ErrInvalidTemplateName) {
		t.Errorf("err = %v, want ErrInvalidTemplateName", err)
	}
}

// Phase L cycle 10: resolved name is the input name (no prefix yet —
// prefix comes from ChainLoader in Phase M).
func TestDirLoader_ResolvedNameIsInput(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "a.txt"), []byte("x"), 0o600)
	loader, _ := NewDirLoader(dir)
	_, resolved, err := loader.Open("a.txt")
	if err != nil {
		t.Fatal(err)
	}
	if resolved != "a.txt" {
		t.Errorf("resolved = %q, want a.txt", resolved)
	}
}

// Phase L cycle 11: non-existent directory fails at construction.
func TestDirLoader_NonExistentDir_ConstructorFails(t *testing.T) {
	t.Parallel()

	_, err := NewDirLoader("/definitely/does/not/exist/xyz123")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// Phase L cycle 12: T3c — NewFSLoader(os.DirFS(...)) escape hatch
// follows symlinks. This documents that users who deliberately want
// to follow symlinks can opt into the non-sandboxed Go primitive.
func TestFSLoader_WithOsDirFS_EscapeHatch(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("symlink semantics differ on Windows")
	}

	outside := t.TempDir()
	target := filepath.Join(outside, "target.txt")
	_ = os.WriteFile(target, []byte("via-symlink"), 0o600)

	root := t.TempDir()
	link := filepath.Join(root, "follow.txt")
	if err := os.Symlink(target, link); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	// os.DirFS deliberately does NOT sandbox symlinks. This is the
	// documented escape hatch for development workflows.
	loader := NewFSLoader(os.DirFS(root))
	src, _, err := loader.Open("follow.txt")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if src != "via-symlink" {
		t.Errorf("src = %q", src)
	}
}
