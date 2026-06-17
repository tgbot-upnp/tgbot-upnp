package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gotd/contrib/http_io"
	"github.com/gotd/contrib/partio"
	"github.com/gotd/contrib/tg_io"
	"github.com/gotd/td/tg"
	"github.com/spf13/viper"
	"github.com/tgbot-upnp/tgbot-upnp/config"
	"github.com/tgbot-upnp/tgbot-upnp/dcpool"
	"github.com/tgbot-upnp/tgbot-upnp/lang"
	"github.com/tgbot-upnp/tgbot-upnp/setup"
	"go.uber.org/zap"
)

// partSize is the chunk size for streaming downloads.
// Telegram supports up to 1 MB per request; larger chunks mean fewer
// HTTP round-trips and higher throughput.
const partSize = 1024 * 1024

type TgVideo struct {
	Doc *tg.Document
	ID  string
}

var tgVideos = make(map[string]*TgVideo)
var httpPort int
var logger *zap.Logger
var downloadPool *dcpool.Pool
var httpServer *http.Server

// SetPool injects the DC connection pool for file downloads.
// Must be called before any ServeFile invocations.
func SetPool(p *dcpool.Pool) {
	downloadPool = p
}

func Server(port int, globalLogger *zap.Logger, translations lang.SetupStrings) {
	httpPort = port
	logger = globalLogger

	setupHandler := setup.Handler(translations)
	mux := http.NewServeMux()
	mux.Handle("/", setupHandler)
	mux.HandleFunc("/video/", video)
	httpServer = &http.Server{
		Addr:        fmt.Sprintf(":%d", httpPort),
		Handler:     mux,
		ReadTimeout: 10 * time.Second,
	}
	go func() {
		logger.Info("http server listening", zap.Int("port", httpPort))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("tgbot-upnp http server failed", zap.String("err", err.Error()))
		}
	}()
}

// Shutdown gracefully stops the HTTP server.
func Shutdown() {
	if httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(ctx)
	}
}
func video(w http.ResponseWriter, r *http.Request) {
	tgVideoId := path.Base(r.URL.Path)
	if tgVideo, ok := tgVideos[tgVideoId]; ok {
		logger.Info("serve video request", zap.String("id", tgVideoId))
		tgVideo.ServeFile(w, r)
	} else {
		logger.Error("Requests for non-existent documents", zap.String("tgVideoId", tgVideoId))
		w.WriteHeader(404)
		_, _ = fmt.Fprintln(w, "File not found")
	}
}

func GetTgVideoPlayUrl(tgVideo *TgVideo, avClientIP net.IP) (url, tgVideoID string, err error) {
	tgVideo.ID = strconv.FormatInt(tgVideo.Doc.AsInputDocumentFileLocation(string(tgVideo.Doc.FileReference)).GetID(), 16)
	tgVideos[tgVideo.ID] = tgVideo

	var prefix string
	if base := viper.GetString(config.BaseURL); base != "" {
		prefix = strings.TrimSuffix(strings.TrimSpace(base), "/")
	} else if host, err := getAvailableHost(avClientIP); err == nil {
		prefix = fmt.Sprintf("http://%s", host)
	} else {
		return "", tgVideo.ID, err
	}

	return fmt.Sprintf("%s/video/%s", prefix, tgVideo.ID), tgVideo.ID, nil
}

func (video *TgVideo) ServeFile(resp http.ResponseWriter, req *http.Request) {
	var client *tg.Client
	if downloadPool != nil {
		client = downloadPool.Client(req.Context(), video.Doc.DCID)
	} else {
		logger.Warn("dcpool not initialized, falling back to default client")
		client = tg.NewClient(nil)
	}

	streamer := partio.NewStreamer(
		tg_io.NewDownloader(client).ChunkSource(
			video.Doc.Size,
			video.Doc.AsInputDocumentFileLocation(string(video.Doc.FileReference)),
		),
		partSize,
	)
	handler := http_io.NewHandler(streamer, video.Doc.Size).
		WithContentType(video.Doc.MimeType)
	handler.ServeHTTP(resp, req)
}

func getAvailableHost(remoteIP net.IP) (string, error) {
	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		switch v := addr.(type) {
		case *net.IPNet:
			if v.Contains(remoteIP) {
				return fmt.Sprintf("%s:%d", v.IP, httpPort), nil
			}
		}
	}
	logger.Error("no available host for upnp")
	return "", errors.New("no available host")
}
