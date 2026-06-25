package setup

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/tgbot-upnp/tgbot-upnp/lang"
	"gopkg.in/yaml.v3"
)

//go:embed setup.html
var setupHTML string

// ---------- GET /api/config ----------

type apiConfigResponse struct {
	Strings map[string]string    `json:"strings"`
	Presets map[string]PresetApp `json:"presets"`
	Config  apiConfigData        `json:"config"`
	Locale  string               `json:"locale"` // e.g. "zh-Hans", "en"
}

type apiConfigData struct {
	CredSource     string `json:"credSource"`
	AppID          string `json:"appId"`
	APIHash        string `json:"apiHash"`
	BotToken       string `json:"botToken"`
	AdminID        string `json:"adminId"`
	AutoAdmin      bool   `json:"autoAdmin"`
	HTTPPort       string `json:"httpPort"`
	BaseURL        string `json:"baseUrl"`
	UserSession    string `json:"userSession"`
	SessionDisplay string `json:"sessionDisplay"`
	SessionStatus  string `json:"sessionStatus"` // "ok" | "unverified" | ""
}

func buildStringsMap(s lang.SetupStrings) map[string]string {
	return map[string]string{
		"title": s.Title, "sub": s.Sub,
		"credSource": s.CredSource, "credBuiltinDesktop": s.CredBuiltinDesktop,
		"credBuiltinTDL": s.CredBuiltinTDL, "credCustom": s.CredCustom,
		"appId": s.AppID, "appIdHint": s.AppIDHint,
		"apiHash": s.APIHash, "apiHashHint": s.APIHashHint,
		"botToken": s.BotToken, "botTokenHint": s.BotTokenHint,
		"adminId": s.AdminID, "adminIdHint": s.AdminIDHint,
		"httpPort": s.HTTPPort,
		"baseUrl":  s.BaseURL, "baseUrlHint": s.BaseURLHint,
		"userSession": s.UserSession, "userSessionHint": s.UserSessionHint,
		"autoAdmin": s.AutoAdmin,
		"btnScan":   s.BtnScan,
		"qrTitle":   s.QRTitle, "qrWaiting": s.QRWaiting, "qrOk": s.QROK,
		"btnSave": s.BtnSave, "saved": s.Saved,
	}
}

func handleAPIConfig(tr lang.SetupStrings) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg := apiConfigData{
			CredSource: DefaultPreset,
			HTTPPort:   "8080",
		}

		// Read existing config.yml
		if raw, err := os.ReadFile("config.yml"); err == nil {
			var c configData
			if yaml.Unmarshal(raw, &c) == nil {
				cfg.AppID = strconv.Itoa(c.AppID)
				cfg.APIHash = c.APIHash
				cfg.BotToken = c.BotToken
				cfg.AdminID = c.AdminID
				if c.HTTPPort > 0 {
					cfg.HTTPPort = strconv.Itoa(c.HTTPPort)
				}
				cfg.BaseURL = c.BaseURL
				cfg.UserSession = c.UserSession
				cfg.AutoAdmin = c.AutoAdmin

				// Determine which preset matches
				matched := false
				for name, p := range Presets {
					if p.AppID == c.AppID {
						cfg.CredSource = name
						matched = true
						break
					}
				}
				if !matched && cfg.AppID != "" {
					cfg.CredSource = "custom"
				}

				// Validate saved session (best-effort)
				if c.UserSession != "" {
					name, uid, ok := ValidateSession(c.AppID, c.APIHash, c.UserSession)
					if ok {
						cfg.SessionDisplay = fmt.Sprintf("%s (ID:%d)", name, uid)
						cfg.SessionStatus = "ok"
					} else {
						cfg.SessionDisplay = "Session saved but could not be verified."
						cfg.SessionStatus = "unverified"
					}
				}
			}
		}

		// Auto-detect base_url from non-localhost requests
		if cfg.BaseURL == "" && !isLocalhost(r.Host) {
			scheme := "http://"
			if r.Header.Get("X-Forwarded-Proto") == "https" {
				scheme = "https://"
			}
			cfg.BaseURL = scheme + r.Host
		}

		resp := apiConfigResponse{
			Strings: buildStringsMap(tr),
			Presets: Presets,
			Config:  cfg,
			Locale:  lang.LocaleSystemTag.String(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

// Handler returns an http.Handler for the configuration page.
// All values are fetched client-side via GET /api/config.
func Handler(translations lang.SetupStrings) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(setupHTML))
	})
	mux.HandleFunc("/api/config", handleAPIConfig(translations))
	mux.HandleFunc("/save", func(w http.ResponseWriter, r *http.Request) {
		handleSaveConfig(w, r)
	})
	mux.HandleFunc("/qrcode", HandleQRCode)
	mux.HandleFunc("/qrcode-status", HandleQRStatus)
	return mux
}

// handleSaveConfig is like handleSave, but does not signal completion.
func handleSaveConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, fmt.Sprintf("parse error: %v", err), http.StatusBadRequest)
		return
	}
	appIDStr := strings.TrimSpace(r.FormValue("app_id"))
	apiHashStr := strings.TrimSpace(r.FormValue("api_hash"))
	appID, err := strconv.Atoi(appIDStr)
	if err != nil || appID == 0 {
		http.Error(w, fmt.Sprintf("invalid app_id: %q", appIDStr), http.StatusBadRequest)
		return
	}
	apiHash := apiHashStr
	botToken := strings.TrimSpace(r.FormValue("bot_token"))
	adminID := strings.TrimSpace(r.FormValue("admin_id"))
	httpPort, _ := strconv.Atoi(strings.TrimSpace(r.FormValue("http_port")))
	if httpPort <= 0 {
		httpPort = 8080
	}
	baseURL := strings.TrimSpace(r.FormValue("base_url"))
	autoAdmin := r.FormValue("auto_admin") == "true"
	if apiHash == "" || botToken == "" || (adminID == "" && !autoAdmin) {
		http.Error(w, "all fields are required", http.StatusBadRequest)
		return
	}
	cfg := configData{AppID: appID, APIHash: apiHash, BotToken: botToken, HTTPPort: httpPort, AdminID: adminID, AutoAdmin: autoAdmin, BaseURL: baseURL, UserSession: strings.TrimSpace(r.FormValue("user_session"))}
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := os.WriteFile("config.yml", data, 0o600); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

type configData struct {
	AppID       int    `yaml:"app_id"`
	APIHash     string `yaml:"api_hash"`
	BotToken    string `yaml:"bot_token"`
	HTTPPort    int    `yaml:"http_port"`
	AdminID     string `yaml:"admin_id"`
	AutoAdmin   bool   `yaml:"auto_admin"`
	BaseURL     string `yaml:"base_url"`
	UserSession string `yaml:"user_session"`
}

// Run starts the browser-based configuration wizard.
// Blocks until the user saves a valid configuration or cancels.
func Run(port int, translations lang.SetupStrings) error {
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return fmt.Errorf("setup: listen: %w", err)
	}

	done := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(setupHTML))
	})
	mux.HandleFunc("/api/config", handleAPIConfig(translations))
	mux.HandleFunc("/save", func(w http.ResponseWriter, r *http.Request) {
		handleSave(w, r, done)
	})
	mux.HandleFunc("/qrcode", HandleQRCode)
	mux.HandleFunc("/qrcode-status", HandleQRStatus)

	srv := &http.Server{
		Addr:        fmt.Sprintf("127.0.0.1:%d", port),
		Handler:     mux,
		ReadTimeout: 10 * time.Second,
	}

	go func() { _ = srv.Serve(listener) }()

	url := fmt.Sprintf("http://127.0.0.1:%d", port)
	if err := OpenBrowser(url); err != nil {
		fmt.Fprintf(os.Stderr, "⚠  Please open %s to configure tgbot-upnp\n", url)
	}

	// Wait for save or Ctrl+C
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-done:
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
		return err
	case <-sig:
		_ = srv.Shutdown(context.Background())
		return fmt.Errorf("setup cancelled by user")
	}
}

// verifyCredentials tries to log in with the given credentials and returns
// the bot's username on success, or an error message on failure.
func handleSave(w http.ResponseWriter, r *http.Request, done chan<- error) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, fmt.Sprintf("parse error: %v", err), http.StatusBadRequest)
		return
	}

	appIDStr := strings.TrimSpace(r.FormValue("app_id"))
	apiHashStr := strings.TrimSpace(r.FormValue("api_hash"))

	appID, err := strconv.Atoi(appIDStr)
	if err != nil || appID == 0 {
		http.Error(w, fmt.Sprintf("invalid app_id: %q", appIDStr), http.StatusBadRequest)
		return
	}
	apiHash := apiHashStr
	botToken := strings.TrimSpace(r.FormValue("bot_token"))
	adminID := strings.TrimSpace(r.FormValue("admin_id"))
	httpPort, _ := strconv.Atoi(strings.TrimSpace(r.FormValue("http_port")))
	if httpPort <= 0 {
		httpPort = 8080
	}
	autoAdmin := r.FormValue("auto_admin") == "true"

	if apiHash == "" || botToken == "" || (adminID == "" && !autoAdmin) {
		http.Error(w, "all fields are required", http.StatusBadRequest)
		return
	}

	baseURL := strings.TrimSpace(r.FormValue("base_url"))
	userSession := strings.TrimSpace(r.FormValue("user_session"))

	cfg := configData{
		AppID:       appID,
		APIHash:     apiHash,
		BotToken:    botToken,
		HTTPPort:    httpPort,
		AdminID:     adminID,
		AutoAdmin:   autoAdmin,
		BaseURL:     baseURL,
		UserSession: userSession,
	}

	data, err := yaml.Marshal(&cfg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := os.WriteFile("config.yml", data, 0o600); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	done <- nil
}

func isLocalhost(host string) bool {
	h := strings.Split(host, ":")[0]
	return h == "127.0.0.1" || h == "localhost" || h == "::1"
}

func OpenBrowser(url string) error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	default:
		return exec.Command("xdg-open", url).Start()
	}
}
