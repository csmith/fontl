package fontl

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
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
	s.mux.HandleFunc("GET /", s.handleIndex)
	s.mux.HandleFunc("GET /api/fonts", s.handleAPIFonts)
	s.mux.HandleFunc("GET /fonts/", s.handleFontServe)
	s.mux.HandleFunc("GET /css/", s.handleCSSServe)
	s.mux.HandleFunc("POST /upload", s.handleUpload)
	s.mux.HandleFunc("POST /edit", s.handleEdit)
	staticFS, _ := fs.Sub(StaticFiles, "static")
	s.mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
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
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
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

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.template.Execute(w, fontList); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("fontFile")
	if err != nil {
		http.Error(w, "Failed to get file from form", http.StatusBadRequest)
		return
	}
	defer file.Close()

	filename := header.Filename
	if !s.storage.IsFontFile(filename) {
		http.Error(w, "Invalid font file type", http.StatusBadRequest)
		return
	}

	fontPath := filepath.Join(s.storage.directory, filename)
	if _, err := os.Stat(fontPath); err == nil {
		http.Error(w, "Font file already exists", http.StatusConflict)
		return
	}

	outFile, err := os.Create(fontPath)
	if err != nil {
		log.Printf("Failed to create font file: %v", err)
		http.Error(w, "Failed to save font file", http.StatusInternalServerError)
		return
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, file); err != nil {
		log.Printf("Failed to copy font file: %v", err)
		http.Error(w, "Failed to save font file", http.StatusInternalServerError)
		return
	}

	fontName := r.FormValue("fontName")
	if fontName == "" {
		fontName = strings.TrimSuffix(filename, filepath.Ext(filename))
	}

	source := r.FormValue("source")
	commercialUse, _ := strconv.ParseBool(r.FormValue("commercialUse"))

	var projects []string
	if projectsStr := r.FormValue("projects"); projectsStr != "" {
		for _, p := range strings.Split(projectsStr, ",") {
			projects = append(projects, strings.TrimSpace(p))
		}
	}

	var tags []string
	if tagsStr := r.FormValue("tags"); tagsStr != "" {
		for _, t := range strings.Split(tagsStr, ",") {
			tags = append(tags, strings.TrimSpace(t))
		}
	}

	metadata := &Metadata{
		Name:          fontName,
		Source:        source,
		CommercialUse: commercialUse,
		Projects:      projects,
		Tags:          tags,
	}

	if err := s.storage.AddUploadedFont(filename, fontPath, metadata); err != nil {
		log.Printf("Failed to add font to storage: %v", err)
		os.Remove(fontPath)
		http.Error(w, "Failed to save font metadata", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully uploaded font: %s", filename)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) handleEdit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	filename := r.FormValue("filename")
	if filename == "" {
		http.Error(w, "Filename is required", http.StatusBadRequest)
		return
	}

	fontName := r.FormValue("fontName")
	if fontName == "" {
		http.Error(w, "Font name is required", http.StatusBadRequest)
		return
	}

	source := r.FormValue("source")
	commercialUse, _ := strconv.ParseBool(r.FormValue("commercialUse"))

	var projects []string
	if projectsStr := r.FormValue("projects"); projectsStr != "" {
		for _, p := range strings.Split(projectsStr, ",") {
			projects = append(projects, strings.TrimSpace(p))
		}
	}

	var tags []string
	if tagsStr := r.FormValue("tags"); tagsStr != "" {
		for _, t := range strings.Split(tagsStr, ",") {
			tags = append(tags, strings.TrimSpace(t))
		}
	}

	metadata := &Metadata{
		Name:          fontName,
		Source:        source,
		CommercialUse: commercialUse,
		Projects:      projects,
		Tags:          tags,
	}

	if err := s.storage.UpdateFontMetadata(filename, metadata); err != nil {
		log.Printf("Failed to update font metadata: %v", err)
		http.Error(w, "Failed to update font metadata", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully updated font metadata: %s", filename)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func initTemplate() *template.Template {
	return template.Must(template.New("index.gotpl").Funcs(template.FuncMap{
		"safeCSS": func(css string) template.CSS {
			return template.CSS(css)
		},
		"hasPrefix": func(s, prefix string) bool {
			return strings.HasPrefix(s, prefix)
		},
		"extractDomain": func(url string) string {
			if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
				// Remove protocol
				url = strings.TrimPrefix(url, "http://")
				url = strings.TrimPrefix(url, "https://")
				// Extract just the domain part (before first slash)
				if idx := strings.Index(url, "/"); idx != -1 {
					url = url[:idx]
				}
				return url
			}
			return url
		},
	}).Parse(IndexTemplate))
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
