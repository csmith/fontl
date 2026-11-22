package fontl

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
)

var ErrFontNotFound = errors.New("font not found")

type Storage struct {
	directory string

	mutex     sync.RWMutex // guards fonts and fontPaths
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
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.fonts = make(map[string]*Metadata)
	m.fontPaths = make(map[string]string)

	return filepath.WalkDir(
		m.directory,
		func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if d.IsDir() || !m.IsFontFile(path) {
				return nil
			}

			if err := m.loadFontMetadata(path); err != nil {
				return fmt.Errorf("failed to load metadata for %s: %w", path, err)
			}

			return nil
		},
	)
}

func (m *Storage) IsFontFile(path string) bool {
	return slices.Contains(
		[]string{".ttf", ".otf", ".woff", ".woff2", ".eot"},
		strings.ToLower(filepath.Ext(path)),
	)
}

func (m *Storage) AddFont(fontPath string, metadata *Metadata) error {
	if !m.IsFontFile(fontPath) {
		return fmt.Errorf("file is not a supported font type: %s", fontPath)
	}

	if _, err := os.Stat(fontPath); os.IsNotExist(err) {
		return fmt.Errorf("font file does not exist: %s", fontPath)
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.fonts[fontPath] = metadata
	return m.saveMetadata(m.getMetadataPath(fontPath), metadata)
}

// GetFonts returns a snapshot of the library, keyed by filename. The returned
// map and metadata must not be modified.
func (m *Storage) GetFonts() map[string]*Metadata {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return maps.Clone(m.fonts)
}

func (m *Storage) GetFontFile(fontName string) (io.ReadCloser, error) {
	m.mutex.RLock()
	fontPath, exists := m.fontPaths[fontName]
	m.mutex.RUnlock()

	if !exists {
		fontPath = filepath.Join(m.directory, fontName)
	}
	return os.Open(fontPath)
}

func (m *Storage) AddUploadedFont(filename, fontPath string, metadata *Metadata) error {
	if !m.IsFontFile(fontPath) {
		return fmt.Errorf("file is not a supported font type: %s", fontPath)
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.fonts[filename] = metadata
	m.fontPaths[filename] = fontPath

	return m.saveMetadata(m.getMetadataPath(fontPath), metadata)
}

func (m *Storage) UpdateFontMetadata(filename string, metadata *Metadata) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.fonts[filename]; !exists {
		return fmt.Errorf("%w: %s", ErrFontNotFound, filename)
	}

	fontPath, exists := m.fontPaths[filename]
	if !exists {
		return fmt.Errorf("font path not found for: %s", filename)
	}

	m.fonts[filename] = metadata
	return m.saveMetadata(m.getMetadataPath(fontPath), metadata)
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
	return filepath.Clean(fontPath) + ".fontl.json"
}

func (m *Storage) saveMetadata(metadataPath string, metadata *Metadata) error {
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(metadataPath, data, 0644)
}
