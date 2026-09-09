package jsonx

import "encoding/json"

const null = "null"

type Field[T any] struct {
	Value T
	Valid bool
}

// Метод с параметризованным ресивером (method with a parameterized receiver). Это не type inference (вывод типов),
// это явное указание type-параметра для метода, чтобы он привязался к дженерик-структуре.
func (f *Field[T]) UnmarshalJSON(data []byte) error {
	f.Valid = true

	if string(data) == null {
		return nil
	}
	return json.Unmarshal(data, &f.Value)
}
