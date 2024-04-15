package template

import (
	"errors"
	"strings"

	"github.com/kaptinlin/filter"
)

// Context stores template variables.
type Context map[string]interface{}

// NewContext creates a Context instance.
func NewContext() Context {
	return make(Context)
}

// Set inserts a variable into the Context, supporting nested keys.
func (c Context) Set(key string, value interface{}) {
	parts := strings.Split(key, ".")
	if len(parts) == 1 {
		c[key] = value
		return
	}

	var current map[string]interface{} = c
	for i, part := range parts {
		if i == len(parts)-1 {
			current[part] = value
		} else {
			if _, exists := current[part]; !exists {
				current[part] = make(map[string]interface{})
			} else if _, ok := current[part].(map[string]interface{}); !ok {
				current[part] = make(map[string]interface{})
			}
			current = current[part].(map[string]interface{})
		}
	}
}

// Get retrieves a variable's value from the Context, supporting nested keys.
func (c Context) Get(key string) (interface{}, error) {
	value, err := filter.Extract(c, key)
	if err != nil {
		switch {
		case errors.Is(err, filter.ErrKeyNotFound):
			return nil, ErrContextKeyNotFound
		case errors.Is(err, filter.ErrInvalidKeyType):
			return nil, ErrContextInvalidKeyType
		case errors.Is(err, filter.ErrIndexOutOfRange):
			return nil, ErrContextIndexOutOfRange
		}

		return nil, err
	}
	return value, nil
}
