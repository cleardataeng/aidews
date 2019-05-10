// Package policy is used to simplify dealing with IAM policies.
//
// IAM policies have several nodes that can be either strings or arrays. For
// example, the "Resource" section can either be a string of a single resources
// ARN, or it can be an array of several resources' ARN strings.
//
// Using this package form marshaling and unmarshaling those policies handles
// such an issue by converting slices of length one, in these places, to strings
// when marshaling. When unmarshaling, strings are converted, conversely, to
// slices of length one.
package policy

import (
	"encoding/json"
	"reflect"
)

// IAMPolicy is an AWS IAM policy document used for converting to and from json.
type IAMPolicy struct {
	Version   string               `yaml:"Version"`
	Statement []IAMPolicyStatement `yaml:"Statement"`
}

// IAMPolicyStatement is a statement object within an AWS IAM policy.
type IAMPolicyStatement struct {
	ID        string                 `json:"Sid" yaml:"Sid"`
	Effect    string                 `yaml:"Effect"`
	Action    StrOrSlice             `yaml:"Action"`
	Resource  StrOrSlice             `json:",omitempty" yaml:"Resource,omitempty"`
	Principal map[string]interface{} `json:",omitempty" yaml:"Principal,omitempty"`
	Condition map[string]interface{} `json:",omitempty" yaml:"Condition,omitempty"`
}

// StrOrSlice is a helper for objects that could be strings or slices.
// In IAM policies, for example, some fields can be strings or arrays.
type StrOrSlice []string

// Equal compares the JSON in two byte slices.
func Equal(a, b []byte) (bool, error) {
	var x, y interface{}
	if err := json.Unmarshal(a, &x); err != nil {
		return false, err
	}
	if err := json.Unmarshal(b, &y); err != nil {
		return false, err
	}
	return reflect.DeepEqual(x, y), nil
}

// MarshalJSON is a method on custom type StringOrSlice that satisfies the
// interface provided by the json package.
// If the length of the given item is only one, we marshal the string. If
// greater than one, we marshal the slice as an array.
func (ss *StrOrSlice) MarshalJSON() ([]byte, error) {
	if len(*ss) == 1 {
		return json.Marshal(([]string)(*ss)[0])
	}
	return json.Marshal(*ss)
}

// UnmarshalJSON is a method on custom type StringOrSlice that satisfies the
// interface provided by the json package.
// If the object in data is of length one we unmarshal into a slice, converting
// the array of one string. If the length is greater than one, we simply
// unmarshal into our slice.
func (ss *StrOrSlice) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		var v []string
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*ss = v
		return nil
	}
	*ss = []string{s}
	return nil
}
