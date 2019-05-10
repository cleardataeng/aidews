package policy

import (
	"github.com/go-yaml/yaml"
	"testing"
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
