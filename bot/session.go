package bot

import (
	"context"
	"os"
	"path/filepath"

	"github.com/gotd/td/session"
)

// fileSession stores session data in a local file.
type fileSession struct {
	path string
}

func newFileSession(path string) *fileSession {
	return &fileSession{path: path}
}

func (s *fileSession) LoadSession(ctx context.Context) ([]byte, error) {
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		// ensure parent directory exists for future writes
		_ = os.MkdirAll(filepath.Dir(s.path), 0o700)
		return nil, nil
	}
	return data, err
}

func (s *fileSession) StoreSession(ctx context.Context, data []byte) error {
	_ = os.MkdirAll(filepath.Dir(s.path), 0o700)
	return os.WriteFile(s.path, data, 0o600)
}

// compile-time check
var _ session.Storage = (*fileSession)(nil)
