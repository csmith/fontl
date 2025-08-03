package fontl

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
)

type Server struct {
	storage  *Storage
	port     int
	mux      *http.ServeMux
	template *template.Template
	server   *http.Server
}

func NewServer(storage *Storage, port int) *Server {
	mux := http.NewServeMux()
	s := &Server{
		storage:  storage,
		port:     port,
		mux:      mux,
		template: initTemplate(),
		server: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: mux,
		},
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	s.mux.HandleFunc("/", s.handleIndex)
	s.mux.HandleFunc("/fonts/", s.handleFontServe)
	s.mux.HandleFunc("/css/", s.handleCSSServe)
}

func (s *Server) handleFontServe(w http.ResponseWriter, r *http.Request) {
	fontName := strings.TrimPrefix(r.URL.Path, "/fonts/")
	if fontName == "" {
		http.Error(w, "Font name required", http.StatusBadRequest)
		return
	}

	file, err := s.storage.GetFontFile(fontName)
	if err != nil {
		http.Error(w, "Font not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	contentType := s.getContentType(fontName)
	w.Header().Set("Content-Type", contentType)

	if _, err := io.Copy(w, file); err != nil {
		log.Printf("Error serving font %s: %v", fontName, err)
	}
}

func (s *Server) handleCSSServe(w http.ResponseWriter, r *http.Request) {
	fontName := strings.TrimPrefix(r.URL.Path, "/css/")
	if fontName == "" {
		http.Error(w, "Font name required", http.StatusBadRequest)
		return
	}

	fonts := s.storage.GetFonts()
	metadata, exists := fonts[fontName]
	if !exists {
		http.Error(w, "Font not found", http.StatusNotFound)
		return
	}

	css := s.generateCSS(fontName, metadata)
	w.Header().Set("Content-Type", "text/css")
	w.Write([]byte(css))
}

func (s *Server) generateCSS(fontName string, metadata *Metadata) string {
	fontFamily := metadata.Name
	if fontFamily == "" {
		fontFamily = strings.TrimSuffix(fontName, filepath.Ext(fontName))
	}

	return fmt.Sprintf(`@font-face {
    font-family: '%s';
    src: url('/fonts/%s');
}`, fontFamily, fontName)
}

type FontData struct {
	Name     string
	Filename string
	Metadata *Metadata
	CSS      string
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	fonts := s.storage.GetFonts()
	var fontList []FontData

	for filename, metadata := range fonts {
		fontFamily := metadata.Name
		if fontFamily == "" {
			fontFamily = strings.TrimSuffix(filename, filepath.Ext(filename))
		}

		fontData := FontData{
			Name:     fontFamily,
			Filename: filename,
			Metadata: metadata,
			CSS:      s.generateCSS(filename, metadata),
		}
		fontList = append(fontList, fontData)
	}

	sort.Slice(fontList, func(i, j int) bool {
		return fontList[i].Name < fontList[j].Name
	})

	w.Header().Set("Content-Type", "text/html")
	if err := s.template.Execute(w, fontList); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func initTemplate() *template.Template {
	return template.Must(template.New("index.gotpl").Funcs(template.FuncMap{
		"safeCSS": func(css string) template.CSS {
			return template.CSS(css)
		},
	}).ParseFiles("index.gotpl"))
}

func (s *Server) getContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".ttf":
		return "font/ttf"
	case ".otf":
		return "font/otf"
	case ".woff":
		return "font/woff"
	case ".woff2":
		return "font/woff2"
	case ".eot":
		return "application/vnd.ms-fontobject"
	default:
		return "application/octet-stream"
	}
}

func (s *Server) Start() error {
	log.Printf("Starting server on %s", s.server.Addr)
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
