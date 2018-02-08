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

    err := decode.Magic(&payload, r,
        decode.JSON,
        decode.ChiRouter,
        decode.QueryParams,
    )

    if err != nil {
        // do something
    }
}
```

For sure you can create shortcuts to common tasks

```go
func decodeGetRequest(container interface{}, r *http.Request) error {
    return decode.Magic(container, r,
        decode.ChiRouter,
        decode.QueryParams,
    )
}

func decodePostRequest(container interface{}, r *http.Request) error {
    return decode.Magic(container, r,
        decode.JSON,
        decode.ChiRouter,
    )
}
```

## Create a custom decoder
Decoder is an abstraction to decode info from a request into a container, container always should be a pointer to a struct

```go
type Decoder func(container interface{}, r *http.Request) error

```
