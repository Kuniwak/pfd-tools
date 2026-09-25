package tools

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteOutput(t *testing.T) {
	testCases := map[string]struct {
		HasDestPath   bool
		UnwritableDir bool
		Content       []byte
		WantWriter    string
		WantFile      string
		WantError     bool
	}{
		"writes to the writer when the dest path is empty": {
			HasDestPath: false,
			Content:     []byte("hello"),
			WantWriter:  "hello",
			WantFile:    "original",
		},
		"writes to the file when the dest path is given": {
			HasDestPath: true,
			Content:     []byte("hello"),
			WantWriter:  "",
			WantFile:    "hello",
		},
		"fails for an unwritable dest path": {
			HasDestPath:   true,
			UnwritableDir: true,
			Content:       []byte("hello"),
			WantWriter:    "",
			WantFile:      "original",
			WantError:     true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			filePath := filepath.Join(t.TempDir(), "out.txt")
			if err := os.WriteFile(filePath, []byte("original"), 0644); err != nil {
				t.Fatal(err)
			}

			destPath := ""
			if testCase.HasDestPath {
				destPath = filePath
			}
			if testCase.UnwritableDir {
				destPath = filepath.Join(filepath.Dir(filePath), "missing-dir", "out.txt")
			}

			buf := bytes.NewBuffer(nil)
			err := WriteOutput(buf, destPath, testCase.Content)
			if testCase.WantError {
				if err == nil {
					t.Fatal("err = nil, want an error")
				}
			} else if err != nil {
				t.Fatalf("WriteOutput: %v", err)
			}

			if buf.String() != testCase.WantWriter {
				t.Errorf("writer = %q, want %q", buf.String(), testCase.WantWriter)
			}

			got, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != testCase.WantFile {
				t.Errorf("file = %q, want %q", string(got), testCase.WantFile)
			}
		})
	}
}

func TestWriteFiles(t *testing.T) {

	type file struct {
		Name    string
		Content string
	}

	testCases := map[string]struct {
		Files []file

		WantFiles map[string]string
		WantError bool
	}{
		"writes every file": {
			Files:     []file{{Name: "a.txt", Content: "A"}, {Name: "b.txt", Content: "B"}, {Name: "c.txt", Content: "C"}},
			WantFiles: map[string]string{"a.txt": "A", "b.txt": "B", "c.txt": "C"},
		},
		"skips a file output with an empty path": {
			Files:     []file{{Name: "", Content: "A"}, {Name: "b.txt", Content: "B"}},
			WantFiles: map[string]string{"b.txt": "B"},
		},
		"fails for an unwritable path": {
			Files:     []file{{Name: filepath.Join("missing-dir", "a.txt"), Content: "A"}},
			WantFiles: map[string]string{},
			WantError: true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()

			files := make([]FileOutput, 0, len(testCase.Files))
			for _, file := range testCase.Files {
				path := ""
				if file.Name != "" {
					path = filepath.Join(dir, file.Name)
				}
				files = append(files, FileOutput{Path: path, Content: []byte(file.Content)})
			}

			err := WriteFiles(files)
			if testCase.WantError {
				if err == nil {
					t.Fatal("err = nil, want an error")
				}
			} else if err != nil {
				t.Fatalf("WriteFiles: %v", err)
			}

			entries, err := os.ReadDir(dir)
			if err != nil {
				t.Fatal(err)
			}
			if len(entries) != len(testCase.WantFiles) {
				t.Errorf("written files = %d, want %d", len(entries), len(testCase.WantFiles))
			}
			for wantName, wantContent := range testCase.WantFiles {
				got, err := os.ReadFile(filepath.Join(dir, wantName))
				if err != nil {
					t.Fatal(err)
				}
				if string(got) != wantContent {
					t.Errorf("%s = %q, want %q", wantName, string(got), wantContent)
				}
			}
		})
	}
}

func TestBuildFileOutput(t *testing.T) {
	buildErr := errors.New("build failed")

	testCases := map[string]struct {
		Path        string
		BuildErr    error
		WantBuilt   bool
		WantPath    string
		WantContent string
		WantError   bool
	}{
		"builds the content for a path": {
			Path:        "out.txt",
			WantBuilt:   true,
			WantPath:    "out.txt",
			WantContent: "content",
		},
		"builds nothing for an empty path": {
			Path:        "",
			WantBuilt:   false,
			WantPath:    "",
			WantContent: "",
		},
		"propagates the build error": {
			Path:      "out.txt",
			BuildErr:  buildErr,
			WantBuilt: true,
			WantError: true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			built := false
			got, err := BuildFileOutput(testCase.Path, func(out io.Writer) error {
				built = true
				if testCase.BuildErr != nil {
					return testCase.BuildErr
				}
				_, writeErr := out.Write([]byte("content"))
				return writeErr
			})

			if built != testCase.WantBuilt {
				t.Errorf("build called = %t, want %t", built, testCase.WantBuilt)
			}
			if testCase.WantError {
				if !errors.Is(err, testCase.BuildErr) {
					t.Fatalf("err = %v, want %v", err, testCase.BuildErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("BuildFileOutput: %v", err)
			}
			if got.Path != testCase.WantPath {
				t.Errorf("Path = %q, want %q", got.Path, testCase.WantPath)
			}
			if string(got.Content) != testCase.WantContent {
				t.Errorf("Content = %q, want %q", string(got.Content), testCase.WantContent)
			}
		})
	}
}
