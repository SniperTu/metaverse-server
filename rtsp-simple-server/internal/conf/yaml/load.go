// Package yaml contains a yaml loader.
package yaml

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v2"
)

func convertKeys(i interface{}) (interface{}, error) {
	switch x := i.(type) {
	case map[interface{}]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			ks, ok := k.(string)
			if !ok {
				return nil, fmt.Errorf("integer keys are not supported (%v)", k)
			}

			var err error
			m2[ks], err = convertKeys(v)
			if err != nil {
				return nil, err
			}
		}
		return m2, nil

	case []interface{}:
		a2 := make([]interface{}, len(x))
		for i, v := range x {
			var err error
			a2[i], err = convertKeys(v)
			if err != nil {
				return nil, err
			}
		}
		return a2, nil
	}

	return i, nil
}

// Load loads the configuration from Yaml.
func Load(buf []byte, dest interface{}) error {
	// load YAML into a generic map
	var temp interface{}
	err := yaml.Unmarshal(buf, &temp)
	if err != nil {
		return err
	}

	// convert interface{} keys into string keys to avoid JSON errors
	temp, err = convertKeys(temp)
	if err != nil {
		return err
	}

	// convert the generic map into JSON
	buf, err = json.Marshal(temp)
	if err != nil {
		return err
	}

	// load JSON into destination
	return json.Unmarshal(buf, dest)
}
