package policy

// CloudFormation templates are often in YAML, so we have a YAML serializer/deserializer.

import (
	"strings"

	yaml "gopkg.in/yaml.v2"
)

// MarshalYAML is a method on custom type StringOrSlice that satisfies the
// interface provided by the yaml package.
// If the length of the given item is only one, we marshal the string. If
// greater than one, we marshal the slice as an array.
func (ss StrOrSlice) MarshalYAML() (interface{}, error) {
	if len(ss) == 1 {
		out, err := yaml.Marshal(([]string)(ss)[0])
		if err != nil {
			return nil, err
		}
		return strings.TrimSpace(string(out)), nil
	}
	return ss, nil
}

// UnmarshalYAML is a method on custom type StringOrSlice that satisfies the
// interface provided by the yaml package.
// If the object in data is of length one we unmarshal into a slice, converting
// the array of one string. If the length is greater than one, we simply
// unmarshal into our slice.
func (ss *StrOrSlice) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		var v []string
		if err := unmarshal(&v); err != nil {
			return err
		}
		*ss = v
		return nil
	}
	*ss = []string{s}
	return nil
}
