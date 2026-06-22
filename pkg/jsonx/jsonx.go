package jsonx

import "encoding/json"

const null = "null"

type Field[T any] struct {
	Value T
	Valid bool
}

func (f *Field[T]) UnmarshalJSON(data []byte) error {
	f.Valid = true

	if string(data) == null {
		return nil
	}
	return json.Unmarshal(data, &f.Value)
}
