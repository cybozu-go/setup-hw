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
                "baz": [ "not ignored, this is not a key" ]
            }
        ]
    },
    "ignored too": {
        "foo": []
    }
}`
	expected := `{"remains":{"foo":[{},{"baz":["not ignored, this is not a key"]}]}}`

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

func TestLeaveFirstItem(t *testing.T) {
	testcases := []struct {
		input    string
		expected []string // multiple possible outputs due to unpredictable order of map items
	}{
		{
			input:    `[ 1, 2, 3 ]`,
			expected: []string{`[1]`},
		},
		{
			input:    `{ "a": [ 1, 2, 3 ], "b": [ 4, 5, 6 ] }`,
			expected: []string{`{"a":[1],"b":[4]}`, `{"b":[4],"a":[1]}`},
		},
		{
			input:    `[ { "a": 1, "b": 2 }, { "c": 3, "d": 4 } ]`,
			expected: []string{`[{"a":1,"b":2}]`, `[{"b":2,"a":1}]`},
		},
		{
			input:    `[ [ 1, 2, 3 ], [ 4, 5, 6 ] ]`,
			expected: []string{`[[1]]`},
		},
	}

OUTER:
	for _, c := range testcases {
		parsed, err := gabs.ParseJSON([]byte(c.input))
		if err != nil {
			t.Fatal(err)
		}
		leaveFirstItem(parsed)
		result := string(parsed.EncodeJSON())
		for _, e := range c.expected {
			if result == e {
				continue OUTER
			}
		}
		t.Errorf("leaveFirstItem() returned unexpected result; expected: one of %s, result: %s", c.expected, result)
	}
}

func TestOmitEmpty(t *testing.T) {
	input := `
{
    "a": {},
    "b": [],
    "c": { "d": {}, "e": [] },
    "f": [ [], {} ],
    "g": { "h": [], "i": 9 },
    "j": [ {}, 10, {}, [ {}, 10.1, {} ] ]
}`
	// multiple possible outputs due to unpredictable order of map items
	expected := []string{
		`{"g":{"i":9},"j":[10,[10.1]]}`,
		`{"j":[10,[10.1]],"g":{"i":9}}`,
	}

	parsed, err := gabs.ParseJSON([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	omitEmpty(parsed)
	result := string(parsed.EncodeJSON())
	for _, e := range expected {
		if result == e {
			return
		}
	}
	t.Errorf("omitEmpty() returned unexpected result; expected: %s, result: %s", expected, result)
}
