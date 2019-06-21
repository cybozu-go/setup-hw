package cmd

import (
	"regexp"
	"testing"

	"github.com/cybozu-go/setup-hw/gabs"
)

func TestRequiredFields(t *testing.T) {
	testcases := []struct {
		input    string
		expected bool
	}{
		{
			input:    `{ "required": "foo" }`,
			expected: true,
		},
		{
			input:    `{ "foo": { "also required": "bar" } }`,
			expected: true,
		},
		{
			input:    `{ "foo": [ { "required too": "bar" } ] }`,
			expected: true,
		},
		{
			input:    `{ "foo": [ "required?  Oops this is not a key." ] }`,
			expected: false,
		},
	}

	pattern := regexp.MustCompile("required")
	for _, c := range testcases {
		parsed, err := gabs.ParseJSON([]byte(c.input))
		if err != nil {
			t.Fatal(err)
		}
		result := requiredFields(parsed, pattern)
		if result != c.expected {
			t.Errorf("requiredFields() returned unexpected result; input: %s, expected: %v, result: %v", c.input, c.expected, result)
		}
	}
}

func TestIgnoreFields(t *testing.T) {
	input := `
{
    "remains": {
        "foo": [
            {
                "ignored": [ "bar" ]
            },
            {
                "bar": [ "not ignored, this is not a key" ]
            }
        ]
    },
    "ignored too": {
        "foo": []
    }
}`
	expected := `{"remains":{"foo":[{},{"bar":["not ignored, this is not a key"]}]}}`

	parsed, err := gabs.ParseJSON([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	pattern := regexp.MustCompile("ignored")
	ignoreFields(parsed, pattern)
	result := string(parsed.EncodeJSON())
	if result != expected {
		t.Errorf("ignoreFields() returned unexpected result; expected: %s, result: %s", expected, result)
	}
}
