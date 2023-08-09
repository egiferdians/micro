package validator

import (
	"errors"
	"fmt"
	"reflect"
)

// V validate struct transformed into array of interface
func V(attributes []interface{}, m interface{}) error {
	var errStr string
	for index, attr := range attributes {
		if attr == "" || attr == nil {
			field := reflect.Indirect(reflect.ValueOf(m)).Type().Field(index).Name
			errStr += fmt.Sprintf("[%s field must has value],", field)
		}
	}
	if errStr == "" {
		return nil
	}
	return errors.New(errStr)
}
