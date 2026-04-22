package workspace

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Workspace defines the boundary for tool file access.
type Workspace struct {
	RootPath string
}

func New(root string) (Workspace, error) {
	resolved, err := filepath.Abs(root)
	if err != nil {
		return Workspace{}, fmt.Errorf("resolve workspace: %w", err)
	}
	return Workspace{RootPath: resolved}, nil
}

func (w Workspace) Resolve(path string) (string, error) {
	candidate := path
	if !filepath.IsAbs(candidate) {
		candidate = filepath.Join(w.RootPath, candidate)
	}

	resolved, err := filepath.Abs(candidate)
	if err != nil {
		return "", fmt.Errorf("resolve path: %w", err)
	}
	if resolved != w.RootPath && !strings.HasPrefix(resolved, w.RootPath+string(filepath.Separator)) {
		return "", fmt.Errorf("path %q escapes workspace", path)
	}
	return resolved, nil
}
