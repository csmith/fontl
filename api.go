package fontl

import (
	"encoding/json"
	"log"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
)

func (s *Server) handleAPIFonts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(fontList); err != nil {
		log.Printf("Error encoding JSON: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
