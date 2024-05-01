package main

import (
	"os"
	"path"
	"path/filepath"
	"strings"
)

type Writer struct {
	dir  string
	sep  string
	ext  string
	out  *os.File
	more bool
}

func (w *Writer) Close() error {
	if w.out != nil && w.out != os.Stdout {
		return w.out.Close()
	}
	return nil
}

func (w *Writer) Write(p []byte) (n int, err error) {
	return w.out.Write(p)
}

func (w *Writer) SwtichFile(name string) error {
	if w.dir != "" {
		if w.out != nil {
			w.out.Close()
			w.out = nil
		}
		if w.ext != "" {
			ext := path.Ext(name)
			name = name[:len(name)-len(ext)] + w.ext
		}

		f, err := os.Create(filepath.Join(w.dir, name))
		if err != nil {
			return err
		}
		w.out = f
	}
	if !w.more {
		w.more = true
	} else if w.dir == "" && w.sep != "" {
		w.out.WriteString(w.sep)
	}
	return nil
}

func newWriter(output string, sep string, ext string) (*Writer, error) {
	if sep != "" {
		sep += "\n"
	}

	outputDir := false
	if output != "-" && !strings.Contains(filepath.Base(output), ".") {
		outputDir = true
	}

	if outputDir {
		err := os.MkdirAll(output, 0755)
		if err != nil {
			return nil, err
		}
		return &Writer{
			dir: output,
			ext: ext,
		}, nil
	}

	var out *os.File
	if output == "-" {
		out = os.Stdout
	} else {
		f, err := os.Create(output)
		if err != nil {
			return nil, err
		}
		out = f
	}
	return &Writer{
		out: out,
		sep: sep,
	}, nil
}
