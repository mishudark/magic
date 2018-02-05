package decode

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// newRwquest is used to generate the request for the various test cases
func newRequest(method, urlStr string, body io.Reader) *http.Request {
	req, _ := http.NewRequest(method, urlStr, body)
	return req
}

func TestQueryParams(t *testing.T) {
	type item struct {
		Name string `path:"name"`
	}

	tc := []struct {
		name      string
		req       *http.Request
		container interface{}
		output    interface{}
	}{
		{
			name:      "query param",
			req:       newRequest("GET", "/foo?name=bob", nil),
			container: &item{},
			output:    &item{Name: "bob"},
		},
		{
			name:      "query param empty",
			req:       newRequest("GET", "/foo", nil),
			container: &item{},
			output:    &item{},
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			QueryParams([]string{"name"})(tt.container, tt.req)
			if diff := cmp.Diff(tt.container, tt.output); diff != "" {
				t.Errorf("%s: -got +want\n%s", tt.name, diff)
			}
		})
	}
}

func TestJSON(t *testing.T) {
	type item struct {
		Name string `json:"name"`
	}

	tc := []struct {
		name      string
		req       *http.Request
		container interface{}
		output    interface{}
		hasErr    bool
	}{
		{
			name:      "empty request",
			req:       newRequest("POST", "/foo", bytes.NewBufferString("")),
			container: &item{},
			output:    &item{},
			hasErr:    true,
		},
		{
			name:      "empty request",
			req:       newRequest("POST", "/foo", nil),
			container: &item{},
			output:    &item{},
			hasErr:    true,
		},
		{
			name:      "empty request",
			req:       newRequest("POST", "/foo", bytes.NewBufferString(`{"name": "foo"}`)),
			container: &item{},
			output:    &item{Name: "foo"},
			hasErr:    false,
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			err := JSON(tt.container, tt.req)
			if diff := cmp.Diff(tt.container, tt.output); diff != "" {
				t.Errorf("%s: -got +want\n%s", tt.name, diff)
			}
			if (tt.hasErr && err == nil) || (!tt.hasErr && err != nil) {
				t.Errorf("%s: expect err to be %t, got: %s", tt.name, tt.hasErr, err)
			}
		})
	}
}

func TestParseToStruct(t *testing.T) {
	type item struct {
		Name   string `form:"name"`
		Number int
		Money  float64
	}

	type itemComplete struct {
		Name    string   `form:"name"`
		Number  int      `form:"number"`
		Money   float64  `form:"money"`
		IsSafe  bool     `form:"issafe"`
		Numbers []int    `form:"numbers"`
		Friends []string `form:"friends"`
	}

	tc := []struct {
		name      string
		container interface{}
		hasErr    bool
		output    interface{}
		form      map[string]string
	}{
		{
			name:      "empty container",
			container: nil,
			hasErr:    true,
			output:    nil,
			form:      map[string]string{},
		},
		{
			name:      "empty form",
			container: nil,
			hasErr:    false,
			output:    nil,
			form:      nil,
		},
		{
			name:      "only string param",
			container: &item{},
			hasErr:    false,
			output: &item{
				Name: "foo",
			},
			form: map[string]string{"name": "foo"},
		},
		{
			name:      "string and number, only name has tag",
			container: &item{},
			hasErr:    false,
			output: &item{
				Name: "foo",
			},
			form: map[string]string{
				"name":   "foo",
				"number": "2",
			},
		},
		{
			name:      "string and number",
			container: &itemComplete{},
			hasErr:    false,
			output: &itemComplete{
				Name:   "foo",
				Number: 2,
			},
			form: map[string]string{
				"name":   "foo",
				"number": "2",
			},
		},
		{
			name:      "string, number and float",
			container: &itemComplete{},
			hasErr:    false,
			output: &itemComplete{
				Name:   "foo",
				Number: 2,
				Money:  12.30,
			},
			form: map[string]string{
				"name":   "foo",
				"number": "2",
				"money":  "12.30",
			},
		},
		{
			name:      "string, number, float and bool",
			container: &itemComplete{},
			hasErr:    false,
			output: &itemComplete{
				Name:   "foo",
				Number: 2,
				Money:  12.30,
				IsSafe: true,
			},
			form: map[string]string{
				"name":   "foo",
				"number": "2",
				"money":  "12.30",
				"issafe": "on",
			},
		},
		{
			name:      "bool is '1'",
			container: &itemComplete{},
			hasErr:    false,
			output: &itemComplete{
				IsSafe: true,
			},
			form: map[string]string{
				"issafe": "1",
			},
		},
		{
			name:      "bool is 'true'",
			container: &itemComplete{},
			hasErr:    false,
			output: &itemComplete{
				IsSafe: true,
			},
			form: map[string]string{
				"issafe": "true",
			},
		},
		{
			name:      "bool is 'yes'",
			container: &itemComplete{},
			hasErr:    false,
			output: &itemComplete{
				IsSafe: true,
			},
			form: map[string]string{
				"issafe": "yes",
			},
		},
		{
			name:      "slice of ints",
			container: &itemComplete{},
			hasErr:    false,
			output: &itemComplete{
				Numbers: []int{1, 2, 3, 4},
			},
			form: map[string]string{
				"numbers": "1,2,3,4",
			},
		},
		{
			name:      "slice of ints with trailing comma",
			container: &itemComplete{},
			hasErr:    true,
			output: &itemComplete{
				Numbers: []int{1, 2, 3, 0},
			},
			form: map[string]string{
				"numbers": "1,2,3,",
			},
		},
		{
			name:      "slice of strings",
			container: &itemComplete{},
			hasErr:    false,
			output: &itemComplete{
				Friends: []string{"Bob", "Carl"},
			},
			form: map[string]string{
				"friends": "Bob,Carl",
			},
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			err := ParseToStruct("form", tt.form, tt.container)
			if diff := cmp.Diff(tt.container, tt.output); diff != "" {
				t.Errorf("%s: -got +want\n%s", tt.name, diff)
			}
			if (tt.hasErr && err == nil) || (!tt.hasErr && err != nil) {
				t.Errorf("%s: expect err to be %t, got: %s", tt.name, tt.hasErr, err)
			}
		})
	}
}
