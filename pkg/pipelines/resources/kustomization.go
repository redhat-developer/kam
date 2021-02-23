package resources

import "sort"

// Kustomization is a structural representation of the Kustomize file format.
type Kustomization struct {
	Resources    []string          `json:"resources,omitempty"`
	Bases        []string          `json:"bases,omitempty"`
	CommonLabels map[string]string `json:"commonLabels,omitempty"`
}

func (k *Kustomization) AddResources(s ...string) {
	k.Resources = removeDuplicatesAndSort(append(k.Resources, s...))
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
