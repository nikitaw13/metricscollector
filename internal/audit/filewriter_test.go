package audit

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_fileWriter_close(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
	}{
		{
			name:     "closes an open file",
			fileName: testFilename,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			newFile, err := openAuditFile(filepath.Join(t.TempDir(), tt.fileName))
			require.NoError(t, err)

			fw := fileWriter{
				file: newFile,
			}

			err = fw.close()
			require.NoError(t, err)

			err = newFile.Close()
			require.ErrorIs(t, err, os.ErrClosed)
		})
	}
}

func Test_openAuditFile(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		wantErr  error
	}{
		{
			name:     "opens a file in an existing directory",
			fileName: testFilename,
			wantErr:  nil,
		},
		{
			name:     "non existing directory",
			fileName: testNonExistingPath,
			wantErr:  os.ErrNotExist,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := openAuditFile(filepath.Join(t.TempDir(), tt.fileName))
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func Test_fileWriter_send(t *testing.T) {
	tests := []struct {
		name          string
		filename      string
		seededContent []byte
		event         []byte
		wantEvent     []byte
		wantErr       error
	}{
		{
			name:          "one event to write",
			filename:      testFilename,
			seededContent: nil,
			event:         testEvent,
			wantEvent:     testComparingEvent,
			wantErr:       nil,
		},
		{
			name:          "two events to write",
			filename:      testFilename,
			seededContent: testEvent,
			event:         testSecondEvent,
			wantEvent:     testComparingDoubleEvent,
			wantErr:       nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			newFile, err := openAuditFile(filepath.Join(t.TempDir(), tt.filename))
			require.NoError(t, err)

			fw := fileWriter{
				file: newFile,
			}

			if tt.seededContent != nil {
				_, err = fmt.Fprintf(newFile, "%s\n", tt.seededContent)
				require.NoError(t, err)
			}

			err = fw.send(tt.event)
			require.ErrorIs(t, err, tt.wantErr)

			if tt.wantErr == nil {
				gotContent, err := os.ReadFile(newFile.Name())
				require.NoError(t, err)
				require.Equal(t, string(tt.wantEvent), string(gotContent))
			}
		})
	}
}
