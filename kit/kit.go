// Package kit contains helpers to proces Go-kit requests
package kit

import (
	"context"
	"fmt"
	"net/http"

	decode "github.com/mishudark/magic-decoder"
)

// DecodeRequestFunc represents the a Go-kit transport http decoder
type Decoder func(context.Context, *http.Request) (response interface{}, err error)

// MagicDecoder is a helper to process incoming Go-kit request
// it should receive a pointer to a struct on arg 'in' and the value of the same
// reference on arg 'out', this is "&user and user", this requirement asumes that
// you expect a value in the output of the decoder
//
// var user User
//
// MagicDecoder(&user, user,
// 	decode.JSON,
// 	decode.ChiRouter,
// 	decode.QueryParams,
// )

func MagicDecoder(in, out interface{}, decoders ...decode.Decoder) Decoder {
	return func(ctx context.Context, r *http.Request) (interface{}, error) {
		err := decode.Magic(in, r,
			decoders...,
		)

		return out, err
	}
}
