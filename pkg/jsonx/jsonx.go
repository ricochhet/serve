package jsonx

import (
	"encoding/json"
	"os"

	"github.com/ricochhet/serve/pkg/errorx"
	"github.com/ricochhet/serve/pkg/fsx"
	"github.com/tidwall/jsonc"
)

// Marshal v of type T.
func Marshal[T any](v T) ([]byte, error) {
	return json.Marshal(v)
}

// Unmarshal data into type T, and store it in store(v).
func Unmarshal[T any](data []byte, store func(T)) error {
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return errorx.WithFrame(err)
	}

	store(v)

	return nil
}

// ReadAndUnmarshal parses a JSON file from the specified path.
func ReadAndUnmarshal[T any](path string) (*T, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errorx.New("os.ReadFile", err)
	}

	var t T
	if err := json.Unmarshal(jsonc.ToJSON(data), &t); err != nil {
		return nil, errorx.New("json.Unmarshal", err)
	}

	return &t, nil
}

// MarshalAndWrite marshales the data to the specified output file.
func MarshalAndWrite[T any](path string, data T) ([]byte, error) {
	b, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return nil, err
	}

	return b, fsx.Write(path, b)
}
