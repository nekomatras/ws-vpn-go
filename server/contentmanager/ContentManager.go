package contentmanager

import (
	"os"
	"net/http"
)

type ContentManager struct {
	page             []byte
	pagePath         string
	staticFolderPath string
}

func New(pagePath string, staticFolderPath string) (*ContentManager, error) {

	manager := &ContentManager{
		pagePath: pagePath,
		staticFolderPath: staticFolderPath,
	}

	page, err := os.ReadFile(pagePath)
	if err != nil {
		return nil, err
	}

	manager.page = page

	return manager, nil
}

func (manager *ContentManager) RegisterHandlers(mux *http.ServeMux) error {

	mux.HandleFunc(manager.pagePath, manager.writeContentHandler)

	fs := http.FileServer(http.Dir(manager.staticFolderPath))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	return nil
}

func (manager *ContentManager) writeContentHandler(w http.ResponseWriter, r *http.Request) {
	manager.WriteContentToResponse(w)
}

func (manager *ContentManager) WriteContentToResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(manager.page)
	w.WriteHeader(http.StatusOK)
}