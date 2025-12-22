package contentmanager

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"ws-vpn-go/common"
)

type ContentManager struct {
	page             []byte
	pagePath         string
	staticFolderPath string

	logger *slog.Logger
}

func New(pagePath string, staticFolderPath string, logger *slog.Logger) (*ContentManager, error) {

	manager := &ContentManager{
		pagePath: pagePath,
		staticFolderPath: staticFolderPath,
		logger: logger,
	}

	page, err := os.ReadFile(pagePath)
	if err != nil {
		return nil, err
	}

	manager.page = page

	return manager, nil
}

func (manager *ContentManager) RegisterHandlers(mux *http.ServeMux) error {

	mux.HandleFunc("/", manager.writeContentHandler)

	fs := http.FileServer(http.Dir(manager.staticFolderPath))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	return nil
}

func (manager *ContentManager) writeContentHandler(w http.ResponseWriter, r *http.Request) {
	manager.WriteContentToResponse(w, r)
}

func (manager *ContentManager) WriteContentToResponse(w http.ResponseWriter, r *http.Request) {
	manager.logger.Warn(fmt.Sprintf("%s try to access %s; Send default content", common.GetRealIP(r), r.URL.String()))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(manager.page)
}