package decode

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi"
)

const (
	formTag = "form"
	pathTag = "path"
)

type Decoder func(container interface{}, r *http.Request) error

func ChiRouter(fields []string) Decoder {
	return func(container interface{}, r *http.Request) error {
		values := make(map[string]string)

		for _, k := range fields {
			values[k] = chi.URLParam(r, k)
		}

		return ParseToStruct(pathTag, values, container)
	}
}

func QueryParams(fields []string) Decoder {
	return func(container interface{}, r *http.Request) error {
		values := make(map[string]string)

		for _, k := range fields {
			values[k] = r.URL.Query().Get(k)
		}

		return ParseToStruct(pathTag, values, container)
	}
}

func JSON(container interface{}, r *http.Request) error {
	if r.Body == nil {
		return errors.New("empty request body")
	}

	return json.NewDecoder(r.Body).Decode(container)
}

func Magic(container interface{}, r *http.Request, decoders ...Decoder) error {
	var err error

	objT := reflect.TypeOf(container)
	if container == nil || !isStructPtr(objT) {
		return fmt.Errorf("%v must be  a struct pointer", container)
	}

	for _, decoder := range decoders {
		if decoder == nil {
			continue
		}

		if err = decoder(container, r); err != nil {
			return err
		}
	}

	return nil
}

const (
	formatTime      = "15:04:05"
	formatDate      = "2006-01-02"
	formatDateTime  = "2006-01-02 15:04:05"
	formatDateTimeT = "2006-01-02T15:04:05"
)

var sliceOfInts = reflect.TypeOf([]int(nil))
var sliceOfStrings = reflect.TypeOf([]string(nil))

func ParseToStruct(structTag string, form map[string]string, container interface{}) error {
	if form == nil {
		return nil
	}

	objT := reflect.TypeOf(container)
	objV := reflect.ValueOf(container)
	if container == nil || !isStructPtr(objT) {
		return fmt.Errorf("%v must be  a struct pointer", container)
	}

	objT = objT.Elem()
	objV = objV.Elem()

	for i := 0; i < objT.NumField(); i++ {
		fieldV := objV.Field(i)
		if !fieldV.CanSet() {
			continue
		}

		fieldT := objT.Field(i)
		if fieldT.Anonymous && fieldT.Type.Kind() == reflect.Struct {
			// err := ParseToStruct(structTag, form, fieldT.Type, fieldV)
			// if err != nil {
			// 	return err
			// }
			continue
		}

		tags := strings.Split(fieldT.Tag.Get(structTag), ",")
		var tag string
		if len(tags) == 0 || len(tags[0]) == 0 {
			continue
		} else if tags[0] == "-" {
			continue
		} else {
			tag = tags[0]
		}

		value := form[tag]
		if value == "" {
			continue
		}

		switch fieldT.Type.Kind() {
		case reflect.Bool:
			switch strings.ToLower(value) {
			case "on", "1", "yes", "true":
				fieldV.SetBool(true)
				continue
			case "off", "0", "no", "false":
				fieldV.SetBool(false)
				continue
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			x, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return err
			}
			fieldV.SetInt(x)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			x, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return err
			}
			fieldV.SetUint(x)
		case reflect.Float32, reflect.Float64:
			x, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return err
			}
			fieldV.SetFloat(x)
		case reflect.Interface:
			fieldV.Set(reflect.ValueOf(value))
		case reflect.String:
			fieldV.SetString(value)
		case reflect.Struct:
			switch fieldT.Type.String() {
			case "time.Time":
				var (
					t   time.Time
					err error
				)
				if len(value) >= 25 {
					value = value[:25]
					t, err = time.ParseInLocation(time.RFC3339, value, time.Local)
				} else if len(value) >= 19 {
					if strings.Contains(value, "T") {
						value = value[:19]
						t, err = time.ParseInLocation(formatDateTimeT, value, time.Local)
					} else {
						value = value[:19]
						t, err = time.ParseInLocation(formatDateTime, value, time.Local)
					}
				} else if len(value) >= 10 {
					if len(value) > 10 {
						value = value[:10]
					}
					t, err = time.ParseInLocation(formatDate, value, time.Local)
				} else if len(value) >= 8 {
					if len(value) > 8 {
						value = value[:8]
					}
					t, err = time.ParseInLocation(formatTime, value, time.Local)
				}
				if err != nil {
					return err
				}
				fieldV.Set(reflect.ValueOf(t))
			}
		case reflect.Slice:
			if fieldT.Type == sliceOfInts {
				formVals := strings.Split(form[tag], ",")
				fieldV.Set(reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(int(1))), len(formVals), len(formVals)))
				for i := 0; i < len(formVals); i++ {
					val, err := strconv.Atoi(formVals[i])
					if err != nil {
						return err
					}
					fieldV.Index(i).SetInt(int64(val))
				}
			} else if fieldT.Type == sliceOfStrings {
				formVals := strings.Split(form[tag], ",")
				fieldV.Set(reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf("")), len(formVals), len(formVals)))
				for i := 0; i < len(formVals); i++ {
					fieldV.Index(i).SetString(formVals[i])
				}
			}
		}
	}
	return nil
}

func isStructPtr(t reflect.Type) bool {
	return t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct
}