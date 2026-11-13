package audit

import (
	"fmt"
	"os"
	"sync"
)

const fileMode = os.FileMode(0644)

// fileWriter appends events to a file.
type fileWriter struct {
	file *os.File
	mu   sync.Mutex
}

// newFileWriter creates a fileWriter that appends events to file.
func newFileWriter(file *os.File) *fileWriter {
	return &fileWriter{
		file: file,
	}
}

// close releases the underlying audit file.
func (fw *fileWriter) close() error {
	return fw.file.Close()
}

// openAuditFile opens fileName for appending, creating it if missing.
func openAuditFile(fileName string) (*os.File, error) {
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, fileMode)
	if err != nil {
		return nil, err
	}
	return file, nil
}

// send appends event as a line to the audit file.
func (fw *fileWriter) send(event []byte) error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	_, err := fmt.Fprintf(fw.file, "%s\n", event)
	if err != nil {
		return err
	}
	return nil
}
