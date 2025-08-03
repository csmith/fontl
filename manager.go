package fontl

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type Storage struct {
	directory string
	fonts     map[string]*Metadata
	fontPaths map[string]string
}

func NewStorage(directory string) *Storage {
	return &Storage{
		directory: directory,
		fonts:     make(map[string]*Metadata),
		fontPaths: make(map[string]string),
	}
}

func (m *Storage) Load() error {
	m.fonts = make(map[string]*Metadata)
	m.fontPaths = make(map[string]string)

	return filepath.WalkDir(m.directory, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if m.isFontFile(path) {
			if err := m.loadFontMetadata(path); err != nil {
				return fmt.Errorf("failed to load metadata for %s: %w", path, err)
			}
		}

		return nil
	})
}

func (m *Storage) isFontFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	fontExts := []string{".ttf", ".otf", ".woff", ".woff2", ".eot"}

	for _, fontExt := range fontExts {
		if ext == fontExt {
			return true
		}
	}

	return false
}

func (m *Storage) loadFontMetadata(fontPath string) error {
	metadataPath := m.getMetadataPath(fontPath)

	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		metadata := &Metadata{
			Name:          "",
			Source:        "",
			CommercialUse: false,
			Projects:      []string{},
			Tags:          []string{},
		}
		filename := filepath.Base(fontPath)
		m.fonts[filename] = metadata
		m.fontPaths[filename] = fontPath
		return m.saveMetadata(metadataPath, metadata)
	}

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return err
	}

	var metadata Metadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return err
	}

	filename := filepath.Base(fontPath)
	m.fonts[filename] = &metadata
	m.fontPaths[filename] = fontPath
	return nil
}

func (m *Storage) getMetadataPath(fontPath string) string {
	dir := filepath.Dir(fontPath)
	filename := filepath.Base(fontPath)
	metadataName := filename + ".fontl.json"
	return filepath.Join(dir, metadataName)
}

func (m *Storage) getDefaultName(fontPath string) string {
	filename := filepath.Base(fontPath)
	ext := filepath.Ext(filename)
	return strings.TrimSuffix(filename, ext)
}

func (m *Storage) saveMetadata(metadataPath string, metadata *Metadata) error {
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(metadataPath, data, 0644)
}

func (m *Storage) AddFont(fontPath string, metadata *Metadata) error {
	if !m.isFontFile(fontPath) {
		return fmt.Errorf("file is not a supported font type: %s", fontPath)
	}

	if _, err := os.Stat(fontPath); os.IsNotExist(err) {
		return fmt.Errorf("font file does not exist: %s", fontPath)
	}

	m.fonts[fontPath] = metadata
	metadataPath := m.getMetadataPath(fontPath)
	return m.saveMetadata(metadataPath, metadata)
}

func (m *Storage) UpdateMetadata(fontPath string, metadata *Metadata) error {
	if _, exists := m.fonts[fontPath]; !exists {
		return fmt.Errorf("font not found in manager: %s", fontPath)
	}

	m.fonts[fontPath] = metadata
	metadataPath := m.getMetadataPath(fontPath)
	return m.saveMetadata(metadataPath, metadata)
}

func (m *Storage) GetFonts() map[string]*Metadata {
	return m.fonts
}

func (m *Storage) GetFontFile(fontName string) (io.ReadCloser, error) {
	fontPath, exists := m.fontPaths[fontName]
	if !exists {
		fontPath = filepath.Join(m.directory, fontName)
	}
	return os.Open(fontPath)
}
