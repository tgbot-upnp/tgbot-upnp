package setup

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gotd/td/clock"
	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth/qrlogin"
	"github.com/gotd/td/telegram/dcs"
	"github.com/gotd/td/tg"
	"github.com/skip2/go-qrcode"
	"go.uber.org/zap"
	"golang.org/x/net/proxy"
)

var (
	qrMu     sync.Mutex
	qrToken  qrlogin.Token
	qrSess   *session.StorageMemory
	qrDone   bool
	qrErr    string
	qrName   string
	qrUserID int64
	qrLog    = zap.NewNop()
)

func SetQRLogger(log *zap.Logger) { qrLog = log }

// ValidateSession tries to use a saved session to get user info.
// Returns name, userID, and whether the session is still valid.
func ValidateSession(appID int, apiHash, userSession string) (name string, userID int64, ok bool) {
	qrLog.Info("validating saved session")
	data, err := base64.StdEncoding.DecodeString(userSession)
	if err != nil {
		qrLog.Warn("session decode failed", zap.Error(err))
		return "", 0, false
	}
	var sess session.StorageMemory
	_ = sess.StoreSession(context.Background(), data)
	client := telegram.NewClient(appID, apiHash, telegram.Options{
		Resolver:       dcs.Plain(dcs.PlainOptions{Dial: proxy.Direct.DialContext}),
		SessionStorage: &sess,
		Clock:          clock.System,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := client.Run(ctx, func(ctx context.Context) error {
		self, err := client.Self(ctx)
		if err != nil {
			return err
		}
		name = self.FirstName
		if self.LastName != "" {
			name += " " + self.LastName
		}
		userID = self.ID
		ok = true
		return nil
	}); err != nil {
		qrLog.Warn("session validation failed", zap.Error(err))
		return "", 0, false
	}
	qrLog.Info("session valid", zap.String("name", name), zap.Int64("id", userID))
	return
}

func HandleQRCode(w http.ResponseWriter, r *http.Request) {
	appID, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("app_id")))
	apiHash := strings.TrimSpace(r.URL.Query().Get("api_hash"))
	if appID == 0 || apiHash == "" {
		http.Error(w, "missing app_id or api_hash", http.StatusBadRequest)
		return
	}

	sess := &session.StorageMemory{}
	dispatcher := tg.NewUpdateDispatcher()
	loggedIn := qrlogin.OnLoginToken(&dispatcher)

	client := telegram.NewClient(appID, apiHash, telegram.Options{
		Resolver:       dcs.Plain(dcs.PlainOptions{Dial: proxy.Direct.DialContext}),
		SessionStorage: sess,
		Clock:          clock.System,
		UpdateHandler:  dispatcher,
	})

	ready := make(chan qrlogin.Token, 1)
	errCh := make(chan error, 1)

	go func() {
		qrLog.Info("qr client started")
		client.Run(context.Background(), func(ctx context.Context) error {
			// client.QR() has built-in DC migration handling
			_, err := client.QR().Auth(ctx, loggedIn,
				func(ctx context.Context, token qrlogin.Token) error {
					qrLog.Info("qr token to show", zap.String("url", token.URL()))
					select {
					case ready <- token:
					default:
					}
					return nil
				},
			)
			if err != nil {
				qrMu.Lock()
				qrErr = err.Error()
				qrDone = true
				qrMu.Unlock()
				return nil
			}
			qrLog.Info("qr login complete")
			// Get user info
			if self, err := client.Self(ctx); err == nil {
				qrMu.Lock()
				qrName = self.FirstName
				if self.LastName != "" {
					qrName += " " + self.LastName
				}
				qrUserID = self.ID
				qrSess = sess
				qrDone = true
				qrMu.Unlock()
			} else {
				qrMu.Lock()
				qrSess = sess
				qrDone = true
				qrMu.Unlock()
			}
			return nil
		})
	}()

	var token qrlogin.Token
	select {
	case t := <-ready:
		token = t
	case err := <-errCh:
		qrMu.Lock()
		if err != nil {
			qrErr = err.Error()
		}
		qrDone = true
		qrMu.Unlock()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	case <-time.After(15 * time.Second):
		http.Error(w, "QR token generation timeout", http.StatusGatewayTimeout)
		return
	}

	qrMu.Lock()
	qrToken = token
	qrSess = sess
	qrDone = false
	qrErr = ""
	qrMu.Unlock()

	img, _ := qrcode.Encode(token.URL(), qrcode.Medium, 256)
	w.Header().Set("Content-Type", "image/png")
	w.Write(img)
}

func HandleQRStatus(w http.ResponseWriter, r *http.Request) {
	qrMu.Lock()
	defer qrMu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	if qrErr != "" {
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": qrErr})
		return
	}
	if qrDone {
		data, _ := qrSess.LoadSession(context.Background())
		session := ""
		if data != nil {
			session = base64.StdEncoding.EncodeToString(data)
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "done",
			"session": session,
			"name":    qrName,
			"userID":  qrUserID,
		})
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "waiting"})
}
