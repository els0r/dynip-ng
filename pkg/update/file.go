package update

import (
	"fmt"
	"html/template"
	"io"
	"os"

	"github.com/els0r/dynip-ng/pkg/cfg"
)

// FileUpdate supplies methods to update the IP in a template and write it to an output file
type FileUpdate struct {
	templatePath string
	outputPath   string
	outputWriter io.Writer
}

type FOption func(*FileUpdate)

func WithOutputWriter(w io.Writer) FOption {
	return func(f *FileUpdate) {
		f.outputWriter = w
	}
}

func NewFileUpdate(cfg *cfg.FileConfig, opts ...FOption) (*FileUpdate, error) {
	f := &FileUpdate{
		templatePath: cfg.Template,
		outputPath:   cfg.Output,
	}

	// apply options
	for _, opt := range opts {
		opt(f)
	}

	if f.outputPath == "" && f.outputWriter == nil {
		return nil, fmt.Errorf("file update must have an output")
	}

	return f, nil
}

func (f *FileUpdate) Update(ip string) error {
	// parse template file
	templ, err := template.ParseFiles(f.templatePath)
	if err != nil {
		return err
	}

	// check if the output is a file
	if f.outputPath != "" {
		fd, err := os.OpenFile(f.outputPath, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			return err
		}
		f.outputWriter = fd
	}

	// execute the template and store result in output
	return templ.Execute(f.outputWriter, ip)
}
