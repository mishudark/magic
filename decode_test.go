package decode

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/mux"
)

// newRwquest is used to generate the request for the various test cases
func newRequest(method, urlStr string, body io.Reader) *http.Request {
	req, _ := http.NewRequest(method, urlStr, body)
	return req
}

func TestQueryParams(t *testing.T) {
	type item struct {
		Name string `form:"name"`
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
			QueryParams(tt.container, tt.req)
			if diff := cmp.Diff(tt.container, tt.output); diff != "" {
				t.Errorf("%s: -got +want\n%s", tt.name, diff)
			}
		})
	}
}

func TestMagic(t *testing.T) {
	type item struct {
		ID    int     `path:"id"`
		Name  string  `form:"name"`
		Pet   string  `form:"pet"`
		Money float64 `json:"money"`
	}

	tc := []struct {
		name      string
		req       *http.Request
		method    string
		route     string
		decoders  []Decoder
		container interface{}
		output    interface{}
		hasErr    bool
	}{
		{
			name:   "POST with route id",
			req:    newRequest("POST", "/foo/2", bytes.NewBufferString(`{"money": 12.34}`)),
			method: "POST",
			route:  "/foo/{id}",
			decoders: []Decoder{
				ChiRouter,
				JSON,
			},
			container: &item{},
			output: &item{
				ID:    2,
				Name:  "",
				Money: 12.34,
			},
			hasErr: false,
		},
		{
			name:   "GET with query param",
			req:    newRequest("GET", "/foo/2?pet=cat", nil),
			method: "GET",
			route:  "/foo/{id:[0-9]+}",
			decoders: []Decoder{
				QueryParams,
				ChiRouter,
			},
			container: &item{},
			output: &item{
				ID:  2,
				Pet: "cat",
			},
			hasErr: false,
		},
		{
			name:   "POST with query param and route id",
			req:    newRequest("GET", "/foo/2?pet=cat&name=bob", bytes.NewBufferString(`{"money": "12.34"}`)),
			method: "POST",
			route:  "/foo/{id:[0-9]+}",
			decoders: []Decoder{
				QueryParams,
				ChiRouter,
				JSON,
			},
			container: &item{},
			output: &item{
				ID:    2,
				Money: 12.34,
				Pet:   "cat",
				Name:  "bob",
			},
			hasErr: false,
		},
		{
			name:   "Nil container",
			req:    newRequest("GET", "/foo/2?pet=cat", nil),
			method: "GET",
			route:  "/foo/{id:[0-9]+}",
			decoders: []Decoder{
				QueryParams,
				ChiRouter,
			},
			container: nil,
			output:    nil,
			hasErr:    true,
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {

			r := chi.NewRouter()
			r.MethodFunc(tt.method, tt.route, func(w http.ResponseWriter, r *http.Request) {
				err := Magic(tt.container, r, tt.decoders...)

				if diff := cmp.Diff(tt.container, tt.output); diff != "" {
					t.Errorf("%s: -got +want\n%s", tt.name, diff)
				}

				if (err == nil) == tt.hasErr {
					t.Errorf("%s: expect err to be %t, got: %s", tt.name, tt.hasErr, err)
				}
			})

			w := httptest.NewRecorder()
			r.ServeHTTP(w, tt.req)
		})

	}
}

func TestChiRouter(t *testing.T) {
	type item struct {
		ID     int `path:"id"`
		TeamID int `path:"team_id"`
	}

	tc := []struct {
		name      string
		route     string
		req       *http.Request
		container interface{}
		output    interface{}
	}{
		{
			name:      "numeric path",
			route:     "/teams/{id}",
			req:       newRequest("GET", "/teams/2", nil),
			container: &item{},
			output:    &item{ID: 2},
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			r.HandleFunc(tt.route, func(w http.ResponseWriter, r *http.Request) {
				ChiRouter(tt.container, r)

				if diff := cmp.Diff(tt.container, tt.output); diff != "" {
					t.Errorf("%s: -got +want\n%s", tt.name, diff)
				}
			})
			w := httptest.NewRecorder()
			r.ServeHTTP(w, tt.req)
		})
	}
}

func TestMuxRouter(t *testing.T) {
	type item struct {
		ID     int `path:"id"`
		TeamID int `path:"team_id"`
	}

	tc := []struct {
		name      string
		route     string
		req       *http.Request
		container interface{}
		output    interface{}
	}{
		{
			name:      "numeric path",
			route:     "/teams/{id}",
			req:       newRequest("GET", "/teams/2", nil),
			container: &item{},
			output:    &item{ID: 2},
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			r := mux.NewRouter()
			r.HandleFunc(tt.route, func(w http.ResponseWriter, r *http.Request) {
				MuxRouter(tt.container, r)

				if diff := cmp.Diff(tt.container, tt.output); diff != "" {
					t.Errorf("%s: -got +want\n%s", tt.name, diff)
				}
			})
			w := httptest.NewRecorder()
			r.ServeHTTP(w, tt.req)
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

			if (err == nil) == tt.hasErr {
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
			if (err == nil) == tt.hasErr {
				t.Errorf("%s: expect err to be %t, got: %s", tt.name, tt.hasErr, err)
			}
		})
	}
}
