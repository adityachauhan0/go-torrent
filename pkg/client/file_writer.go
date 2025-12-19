package client

import (
	"os"
	"path/filepath"
)

// FileWriter handles writing piece data to the output directory.
type FileWriter struct {
	OutputDir string
}

func NewFileWriter(outputDir string) *FileWriter {
	return &FileWriter{
		OutputDir: outputDir,
	}
}

// Write writes the data to the specified file path (relative to output dir) at offset.
func (fw *FileWriter) Write(path string, offset int64, data []byte) error {
	fullPath := filepath.Join(fw.OutputDir, path)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	f, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteAt(data, offset)
	return err
}

// Read reads data from the specified file path at offset.
func (fw *FileWriter) Read(path string, offset int64, length int) ([]byte, error) {
	fullPath := filepath.Join(fw.OutputDir, path)
	f, err := os.Open(fullPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	buf := make([]byte, length)
	_, err = f.ReadAt(buf, offset)
	return buf, err
}
