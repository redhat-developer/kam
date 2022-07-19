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

// From https://github.com/redhat-developer/kam/tree/master/pkg/pipelines/resources

package resources

import "sort"

// Kustomization is a structural representation of the Kustomize file format.
type Kustomization struct {
	APIVersion   string            `json:"apiVersion,omitempty"`
	Kind         string            `json:"kind,omitempty"`
	Resources    []string          `json:"resources,omitempty"`
	Bases        []string          `json:"bases,omitempty"`
	CommonLabels map[string]string `json:"commonLabels,omitempty"`
}

func (k *Kustomization) AddResources(s ...string) {
	k.Resources = removeDuplicatesAndSort(append(k.Resources, s...))
}

func (k *Kustomization) AddBases(s ...string) {
	k.Bases = removeDuplicatesAndSort(append(k.Bases, s...))
}

func removeDuplicatesAndSort(s []string) []string {
	exists := make(map[string]bool)
	out := []string{}
	for _, v := range s {
		if !exists[v] {
			out = append(out, v)
			exists[v] = true
		}
	}
	sort.Strings(out)
	return out
}
