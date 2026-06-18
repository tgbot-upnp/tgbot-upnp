package setup

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
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

// Use [[ ]] as template delimiters so that {{.Username}} and {{.Error}}
// in translation strings (intended for JS runtime replacement) are not
// consumed by Go's template engine.
var tpl = template.Must(
	template.New("setup").
		Delims("[[", "]]").
		Funcs(template.FuncMap{
			"json": func(v any) template.JS {
				b, _ := json.Marshal(v)
				return template.JS(b)
			},
		}).
		Parse(setupHTML),
)

type pageData struct {
	Title              string
	Sub                string
	CredSource         string
	CredBuiltinDesktop string
	CredBuiltinTDL     string
	CredCustom         string
	AppID              string
	AppIDHint          string
	APIHash            string
	APIHashHint        string
	BotToken           string
	BotTokenHint       string
	AdminID            string
	AdminIDHint        string
	HTTPPort           string
	BaseURL            string
	BaseURLHint        string
	BtnSave            string
	PresetsJSON        template.JS
	Saved              template.JS
	// Pre-filled config values (empty if no config exists)
	AppIDVal    string
	APIHashVal  string
	BotTokenVal string
	AdminIDVal  string
	HTTPPortVal string
	BaseURLVal  string
	SavedMsg    template.JS
}

func pageDataFromStrings(s lang.SetupStrings) pageData {
	presetsJSON, _ := json.Marshal(map[string]PresetApp{
		"desktop": Presets["desktop"],
		"tdl":     Presets["tdl"],
	})
	return pageData{
		Title:              s.Title,
		Sub:                s.Sub,
		CredSource:         s.CredSource,
		CredBuiltinDesktop: s.CredBuiltinDesktop,
		CredBuiltinTDL:     s.CredBuiltinTDL,
		CredCustom:         s.CredCustom,
		AppID:              s.AppID,
		AppIDHint:          s.AppIDHint,
		APIHash:            s.APIHash,
		APIHashHint:        s.APIHashHint,
		BotToken:           s.BotToken,
		BotTokenHint:       s.BotTokenHint,
		AdminID:            s.AdminID,
		AdminIDHint:        s.AdminIDHint,
		HTTPPort:           s.HTTPPort,
		BaseURL:            s.BaseURL,
		BaseURLHint:        s.BaseURLHint,
		BtnSave:            s.BtnSave,
		PresetsJSON:        template.JS(presetsJSON),
		Saved:              template.JS(jsonString(s.Saved)),
		HTTPPortVal:        "8080",
	}
}

func jsonString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

// Handler returns an http.Handler that serves the configuration page.
// If a config file exists, values are pre-filled. The save endpoint
// writes to the config file without shutting down the server.
func Handler(translations lang.SetupStrings) http.Handler {
	data := pageDataFromStrings(translations)

	// Pre-fill from existing config file
	if cfg, err := os.ReadFile("config.yml"); err == nil {
		var c configData
		if yaml.Unmarshal(cfg, &c) == nil {
			data.AppIDVal = strconv.Itoa(c.AppID)
			data.APIHashVal = c.APIHash
			data.BotTokenVal = c.BotToken
			data.AdminIDVal = c.AdminID
			if c.HTTPPort > 0 {
				data.HTTPPortVal = strconv.Itoa(c.HTTPPort)
			} else {
				data.HTTPPortVal = "8080"
			}
			data.BaseURLVal = c.BaseURL
		}
	}

	// Use a modified save message for reconfiguration
	data.SavedMsg = template.JS(jsonString(translations.Saved))

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Auto-suggest base_url from request Host if not localhost
		if data.BaseURLVal == "" && !isLocalhost(r.Host) {
			// Host includes port, e.g. "192.168.1.100:8080"
			if scheme := r.Header.Get("X-Forwarded-Proto"); scheme == "https" {
				data.BaseURLVal = "https://" + r.Host
			} else {
				data.BaseURLVal = "http://" + r.Host
			}
		}
		_ = tpl.Execute(w, data)
	})
	mux.HandleFunc("/save", func(w http.ResponseWriter, r *http.Request) {
		handleSaveConfig(w, r, translations)
	})
	return mux
}

// handleSaveConfig is like handleSave, but does not signal completion.
func handleSaveConfig(w http.ResponseWriter, r *http.Request, tr lang.SetupStrings) {
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
	if apiHash == "" || botToken == "" || adminID == "" {
		http.Error(w, "all fields are required", http.StatusBadRequest)
		return
	}
	cfg := configData{AppID: appID, APIHash: apiHash, BotToken: botToken, HTTPPort: httpPort, AdminID: adminID, BaseURL: baseURL}
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
	AppID    int    `yaml:"app_id"`
	APIHash  string `yaml:"api_hash"`
	BotToken string `yaml:"bot_token"`
	HTTPPort int    `yaml:"http_port"`
	AdminID  string `yaml:"admin_id"`
	BaseURL  string `yaml:"base_url"`
}

// Run starts the browser-based configuration wizard on the given port.
// Blocks until the user saves a valid configuration or cancels.
func Run(port int, translations lang.SetupStrings) error {
	data := pageDataFromStrings(translations)

	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return fmt.Errorf("setup: listen: %w", err)
	}

	done := make(chan error, 1)

	// Handlers that use the localized templates from the test/save responses
	tr := translations // capture for handlers

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if !isLocalhost(r.Host) {
			scheme := "http://"
			if r.Header.Get("X-Forwarded-Proto") == "https" {
				scheme = "https://"
			}
			data.BaseURLVal = scheme + r.Host
		}
		_ = tpl.Execute(w, data)
	})
	mux.HandleFunc("/save", func(w http.ResponseWriter, r *http.Request) {
		handleSave(w, r, done, tr)
	})

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
func handleSave(w http.ResponseWriter, r *http.Request, done chan<- error, tr lang.SetupStrings) {
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

	if apiHash == "" || botToken == "" || adminID == "" {
		http.Error(w, "all fields are required", http.StatusBadRequest)
		return
	}

	baseURL := strings.TrimSpace(r.FormValue("base_url"))

	cfg := configData{
		AppID:    appID,
		APIHash:  apiHash,
		BotToken: botToken,
		HTTPPort: httpPort,
		AdminID:  adminID,
		BaseURL:  baseURL,
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
