package tools

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

type FileOutput struct {
	Path    string
	Content []byte
}

func WriteOutput(w io.Writer, destPath string, out []byte) error {
	if destPath != "" {
		return WriteFiles([]FileOutput{{Path: destPath, Content: out}})
	}
	if _, err := w.Write(out); err != nil {
		return fmt.Errorf("tools.WriteOutput: %w", err)
	}
	return nil
}

func WriteFiles(files []FileOutput) error {
	for _, file := range files {
		if file.Path == "" {
			continue
		}
		if err := os.WriteFile(file.Path, file.Content, 0644); err != nil {
			return fmt.Errorf("tools.WriteFiles: writing %s: %w", file.Path, err)
		}
	}
	return nil
}

func BuildFileOutput(path string, build func(out io.Writer) error) (FileOutput, error) {
	if path == "" {
		return FileOutput{}, nil
	}
	buf := bytes.NewBuffer(nil)
	if err := build(buf); err != nil {
		return FileOutput{}, fmt.Errorf("tools.BuildFileOutput: building %s: %w", path, err)
	}
	return FileOutput{Path: path, Content: buf.Bytes()}, nil
}
