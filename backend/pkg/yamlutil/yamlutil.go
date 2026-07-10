package yamlutil

import (
	"sigs.k8s.io/yaml"
)

// MarshalWithoutManagedFields marshals a K8s object to YAML,
// removing metadata.managedFields for cleaner display.
func MarshalWithoutManagedFields(obj interface{}) (string, error) {
	data, err := yaml.Marshal(obj)
	if err != nil {
		return "", err
	}

	// Parse into generic map to remove managedFields
	var m map[string]interface{}
	if err := yaml.Unmarshal(data, &m); err != nil {
		// If parsing fails, return original YAML
		return string(data), nil
	}

	// Remove managedFields from metadata
	if metadata, ok := m["metadata"].(map[string]interface{}); ok {
		delete(metadata, "managedFields")
	}

	result, err := yaml.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(result), nil
}
