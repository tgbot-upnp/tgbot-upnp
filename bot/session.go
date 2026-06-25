package bot

import (
	"context"
	"os"
	"path/filepath"

	"github.com/gotd/td/session"
	"go.uber.org/zap"
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
		_ = os.MkdirAll(filepath.Dir(s.path), 0o700)
		botLogger.Info("session file not found, will create new", zap.String("path", s.path))
		return nil, nil
	}
	if err != nil {
		botLogger.Warn("session load error", zap.String("path", s.path), zap.Error(err))
	} else {
		botLogger.Info("session loaded", zap.String("path", s.path), zap.Int("size", len(data)))
	}
	return data, err
}

func (s *fileSession) StoreSession(ctx context.Context, data []byte) error {
	_ = os.MkdirAll(filepath.Dir(s.path), 0o700)
	if err := os.WriteFile(s.path, data, 0o600); err != nil {
		botLogger.Warn("session store failed", zap.String("path", s.path), zap.Error(err))
		return err
	}
	botLogger.Info("session stored", zap.String("path", s.path), zap.Int("size", len(data)))
	return nil
}

// compile-time check
var _ session.Storage = (*fileSession)(nil)
