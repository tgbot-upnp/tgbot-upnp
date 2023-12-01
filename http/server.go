package http

import (
	"errors"
	"fmt"
	"github.com/gotd/contrib/http_io"
	"github.com/gotd/contrib/partio"
	"github.com/gotd/contrib/tg_io"
	"github.com/gotd/td/tg"
	"go.uber.org/zap"
	"net"
	"net/http"
	"path"
	"strconv"
	"time"
)

type TgVideo struct {
	Api *tg.Client
	ID  string
	Doc *tg.Document
}

var tgVideos = make(map[string]*TgVideo)
var httpPort int
var logger *zap.Logger

func Server(port int, globalLogger *zap.Logger) {
	httpPort = port
	logger = globalLogger
	mux := http.NewServeMux()
	mux.HandleFunc("/", index)
	mux.HandleFunc("/video/", video)
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", httpPort),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			logger.Fatal("gbot-upnp http server failed", zap.String("err", err.Error()))
		}
	}()
	logger.Info("tgbot-upnp http server has been started...", zap.Int("port", httpPort))
}
func index(w http.ResponseWriter, r *http.Request) {
	_, err := fmt.Fprintln(w, "tgbot-upnp http server")
	if err != nil {
		logger.Error("tgbot-upnp http server failed", zap.String("err", err.Error()))
	}
}
func video(w http.ResponseWriter, r *http.Request) {
	tgVideoId := path.Base(r.URL.Path)
	if tgVideo, ok := tgVideos[tgVideoId]; ok {
		tgVideo.ServeFile(w, r)
	} else {
		logger.Error("Requests for non-existent documents", zap.String("tgVideoId", tgVideoId))
		w.WriteHeader(404)
		_, _ = fmt.Fprintln(w, "File not found")
	}
}

func GetTgVideoPlayUrl(tgVideo *TgVideo, avClientIP net.IP) (url, tgVideoID string, err error) {
	tgVideo.ID = strconv.FormatInt(tgVideo.Doc.AsInputDocumentFileLocation().GetID(), 16)
	tgVideos[tgVideo.ID] = tgVideo
	if httpServer, err := getAvailableHost(avClientIP); err == nil {
		return fmt.Sprintf("http://%s/video/%s", httpServer, tgVideo.ID), tgVideo.ID, nil
	} else {
		return "", tgVideo.ID, err
	}
}

func (video *TgVideo) ServeFile(resp http.ResponseWriter, req *http.Request) {
	streamer := partio.NewStreamer(tg_io.NewDownloader(video.Api).ChunkSource(video.Doc.GetSize(), video.Doc.AsInputDocumentFileLocation()), 128*1024)
	handler := http_io.NewHandler(streamer, video.Doc.GetSize()).WithContentType(video.Doc.GetMimeType())
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
