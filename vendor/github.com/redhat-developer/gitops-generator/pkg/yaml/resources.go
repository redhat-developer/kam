//
// Copyright 2021-2022 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// From https://github.com/redhat-developer/kam/tree/master/pkg/pipelines/yaml
package yaml

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/afero"
	"sigs.k8s.io/yaml"
)

// From https://github.com/redhat-developer/kam/blob/master/pkg/pipelines/yaml/resources.go

// WriteResources takes a prefix path, and a map of paths to values, and will
// marshal the values to the filenames as YAML resources, joining the prefix to
// the filenames before writing.
//
// It returns the list of filenames written out.
func WriteResources(fs afero.Fs, path string, files map[string]interface{}) ([]string, error) {
	path, err := homedir.Expand(path)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve path to file: %v", err)
	}
	filenames := make([]string, 0)
	for filename, item := range files {
		err := MarshalItemToFile(fs, filepath.Join(path, filename), item)
		if err != nil {
			return nil, err
		}
		filenames = append(filenames, filename)
	}
	return filenames, nil
}

// MarshalItemToFile marshals item to file
func MarshalItemToFile(fs afero.Fs, filename string, item interface{}) error {
	err := fs.MkdirAll(filepath.Dir(filename), 0755)
	if err != nil {
		return fmt.Errorf("failed to MkDirAll for %s: %v", filename, err)
	}
	f, err := fs.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to Create file %s: %v", filename, err)
	}
	defer f.Close()
	return MarshalOutput(f, item)
}

// MarshalOutput marshal output to given writer
func MarshalOutput(out io.Writer, output interface{}) error {
	data, err := yaml.Marshal(output)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %v", err)
	}
	_, err = fmt.Fprintf(out, "%s", data)
	if err != nil {
		return fmt.Errorf("failed to write data: %v", err)
	}
	return nil
}
