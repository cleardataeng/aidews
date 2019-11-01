package policy

import (
	"testing"

	yaml "gopkg.in/yaml.v2"
)

func TestMarshalYAML(t *testing.T) {
	res := &struct {
		SS0 StrOrSlice
		SS1 StrOrSlice
	}{
		SS0: StrOrSlice([]string{"test"}),
		SS1: StrOrSlice([]string{"test", "slice"}),
	}
	b, err := yaml.Marshal(res)
	if err != nil {
		t.Error(err)
	}
	want := `ss0: test
ss1:
- test
- slice
`

	if string(b) != want {
		t.Errorf("incorrectly marshaled; want: %s, got: %s", want, string(b))
	}
}

func TestUnmarshalYAML(t *testing.T) {

	type MyType struct {
		Empty   StrOrSlice `yaml:"empty"`
		Str     StrOrSlice `yaml:"str"`
		List    StrOrSlice `yaml:"list"`
		Errored StrOrSlice
	}

	input := `
empty:
str: "foo"
list:
  - bar
  - baz
errored:
  foo: bar
`
	output := &MyType{}
	want := "yaml: unmarshal errors:\n  line 8: cannot unmarshal !!map into []string"
	if err := yaml.Unmarshal([]byte(input), output); err == nil {
		t.Errorf("expected error; want: %s", err)
	} else {
		if err.Error() != want {
			t.Errorf("unexpected error; want: %s, got: %s", want, err)
		}
	}

	if output.Empty != nil {
		t.Error("Empty was not empty")
	}
	if output.Str[0] != "foo" {
		t.Errorf("Str wanted %s got %s", "foo", output.Str[0])
	}
	if output.List[1] != "baz" {
		t.Errorf("List wanted %s got %s", "baz", output.List[1])
	}
}

func TestIAMPolicyYAML(t *testing.T) {
	policy := IAMPolicy{
		Version: "12",
	}
	out, err := yaml.Marshal(policy)
	if err != nil {
		t.Errorf("unexpected error; got: %s", err)
	}
	want := "Version: \"12\"\nStatement: []\n"
	if string(out) != want {
		t.Errorf("unexpected out; want: %s got %s", want, string(out))
	}
}

func TestIAMPolicyStatementEmptyYAML(t *testing.T) {
	policy := IAMPolicyStatement{
		ID: "12",
	}
	out, err := yaml.Marshal(policy)
	if err != nil {
		t.Errorf("unexpected error; got: %s", err)
	}
	want := `Sid: "12"
Effect: ""
Action: []
`
	if string(out) != want {
		t.Errorf("unexpected out; want: %s got %s", want, string(out))
	}
}

func TestIAMPolicyStatementNotEmptyYAML(t *testing.T) {
	policy := IAMPolicyStatement{
		ID:        "12",
		Effect:    "allow",
		Action:    StrOrSlice{"iam:Login"},
		Resource:  StrOrSlice{"iam:"},
		Principal: map[string]StrOrSlice{"AWS": StrOrSlice{"iam"}},
		Condition: map[string]interface{}{"String": "matching ARN"},
	}
	out, err := yaml.Marshal(policy)
	if err != nil {
		t.Errorf("unexpected error; got: %s", err)
	}
	want := `Sid: "12"
Effect: allow
Action: iam:Login
Resource: '''iam:'''
Principal:
  AWS: iam
Condition:
  String: matching ARN
`
	if string(out) != want {
		t.Errorf("unexpected out; want: %s got %s", want, string(out))
	}
}

func TestIAMPolicyStatementUnmarshalYAML(t *testing.T) {
	in := `Sid: "12"
Effect: allow
Action: iam:Login
Resource: '''iam:'''
Principal:
  AWS: iam
Condition:
  String: matching ARN
`
	policy := IAMPolicyStatement{}
	err := yaml.Unmarshal(([]byte)(in), &policy)
	if err != nil {
		t.Errorf("unexpected error; got: %s", err)
	}
	if policy.ID != "12" {
		t.Errorf("want: %s got: %s", "12", policy.ID)
	}
	if policy.Action[0] != "iam:Login" {
		t.Errorf("want: %s got: %s", "iam:Login", policy.Action)
	}
}
