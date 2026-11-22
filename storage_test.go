package fontl

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStorage_ConcurrentReadsAndWrites(t *testing.T) {
	dir := t.TempDir()
	storage := NewStorage(dir)
	require.NoError(t, storage.Load())

	var wg sync.WaitGroup
	for i := range 20 {
		filename := fmt.Sprintf("font%d.ttf", i)
		fontPath := filepath.Join(dir, filename)
		require.NoError(t, os.WriteFile(fontPath, nil, 0644))

		wg.Go(
			func() {
				assert.NoError(t, storage.AddUploadedFont(filename, fontPath, &Metadata{}))
				assert.NoError(t, storage.UpdateFontMetadata(filename, &Metadata{Name: filename}))
			},
		)
		wg.Go(
			func() {
				for name, metadata := range storage.GetFonts() {
					_ = name + metadata.Name
				}
			},
		)
	}
	wg.Wait()

	assert.Len(t, storage.GetFonts(), 20)
}
