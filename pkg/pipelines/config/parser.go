package config

import (
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"

	"github.com/spf13/afero"
	"sigs.k8s.io/yaml"
)

// Parse decodes YAML describing an environment manifest.
func Parse(in io.Reader) (*Manifest, error) {
	m := &Manifest{}
	buf, err := ioutil.ReadAll(in)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(buf, m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// ParseFile is a wrapper around Parse that accepts a filename, it opens and
// parses the file, and closes it.
func ParseFile(fs afero.Fs, filename string) (*Manifest, error) {
	f, err := fs.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Parse(f)
}

// ParsePipelinesFolder will accept the pipelines folder path
// and appends pipelines file name before parsing it
func ParsePipelinesFolder(fs afero.Fs, folderPath string) (*Manifest, error) {
	info, err := fs.Stat(folderPath)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("The path %q is a file path (required directory path)", folderPath)
	}
	return ParseFile(fs, filepath.Join(folderPath, PipelinesFile))
}
