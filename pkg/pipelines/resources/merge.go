package resources

// Resources represents a set of filename -> Go struct with the filenames as
// keys, and the values are values to be serialized to YAML.
type Resources map[string]interface{}

// Merge merges a set of resources in from to the set of resources in to,
// replacing existing keys and returns a new set of resources.
func Merge(from, to Resources) Resources {
	merged := Resources{}
	for k, v := range to {
		merged[k] = v
	}
	for k, v := range from {
		merged[k] = v
	}
	return merged
}
