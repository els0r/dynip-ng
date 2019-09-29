package update

import (
	"fmt"
	"html/template"
	"io"
	"os"

	"github.com/els0r/dynip-ng/pkg/cfg"
	"github.com/els0r/dynip-ng/pkg/logging"
	log "github.com/els0r/log"
)

// FileUpdate supplies methods to update the IP in a template and write it to an output file
type FileUpdate struct {
	templatePath      string
	outputPath        string
	outputWriteCloser io.WriteCloser
	log               log.Logger
}

// FOption can be used to configure optional parameters for the file updater
type FOption func(*FileUpdate)

// WithOutputWriteCloser allows for a more generic way of writing the output
func WithOutputWriteCloser(wc io.WriteCloser) FOption {
	return func(f *FileUpdate) {
		f.outputWriteCloser = wc
	}
}

// NewFileUpdate creates a file updater
func NewFileUpdate(cfg *cfg.FileConfig, opts ...FOption) (*FileUpdate, error) {
	f := &FileUpdate{
		templatePath: cfg.Template,
		outputPath:   cfg.Output,
		log:          logging.Get(),
	}

	// apply options
	for _, opt := range opts {
		opt(f)
	}

	if f.outputPath == "" && f.outputWriteCloser == nil {
		return nil, fmt.Errorf("file update must have an output")
	}

	return f, nil
}

// Name returns a human-readable identifier for the updater
func (f *FileUpdate) Name() string {
	return "file updater"
}

// Update takes the IP and writes it to the specified output file using the provided
// input template
func (f *FileUpdate) Update(ip string) error {
	f.log.Debugf("updating file: %s", f.outputPath)

	// parse template file
	templ, err := template.ParseFiles(f.templatePath)
	if err != nil {
		return err
	}

	// check if the output is a file
	if f.outputPath != "" {
		fd, err := os.OpenFile(f.outputPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			return err
		}
		f.outputWriteCloser = fd
	}
	defer func(wc io.WriteCloser) {
		wc.Close()
	}(f.outputWriteCloser)

	// execute the template and store result in output
	return templ.Execute(f.outputWriteCloser, ip)
}
