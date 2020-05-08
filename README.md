# magic-decoder
This lib makes your http decoders "magic"

## install 
`go get -u github.com/mishudark/magic-decoder`

## Usage
Define the struct as usual, `path` and `form` tags are available, to decode fields from query params, and router, this will convert to the struct field type magically

```go
type item struct {
    ID     int     `path:"id"`
    TeamID int     `path:"team_id"`
    SDRN   string  `form:"sdrn"`
    Amount float64 `json:"amount"`
    Name   string  `json:"name"`
}
```

### list of decoders

```go
JSON - "json" tag
ChiRouter - "path" tag
MuxRouter - "path" tag
QueryParams - "form" tag
```

Use the needed decoders with `decode.Magic`

```go
func magicHandler(w http.ResponseWriter, r *http.Request) {
    var payload item

    err := magic.Decode(&payload, r,
        magic.JSON,
        magic.ChiRouter,
        magic.QueryParams,
    )

    if err != nil {
        // do something
    }
}
```

For sure you can create shortcuts to common tasks

```go
func decodeGetRequest(r *http.Request, container interface{}) error {
    return decode.Decode(r, container,
        magic.ChiRouter,
        magic.QueryParams,
    )
}

func decodePostRequest(r *http.Request, container interface{}) error {
    return magic.Decode(r, container,
        magic.JSON,
        magic.ChiRouter,
    )
}
```

## Create a custom decoder
Decoder is an abstraction to decode info from a request into a container, container always should be a pointer to a struct

```go
type Decoder func(r *http.Request, container interface{}) error

```
