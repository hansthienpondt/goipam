package encoding

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Attribute struct {
	Key, Value string
}

func (a *Attribute) String() string {
	return fmt.Sprintf("%s=%s", a.Key, a.Value)
}

func (a *Attribute) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{
		a.Key: a.Value,
	})
}

type Attributes []Attribute

func (a *Attributes) Attributes() []Attribute {
	return *a
}

func (a *Attributes) String() string {
	var result []string
	for _, v := range a.Attributes() {
		result = append(result, v.String())
	}
	return fmt.Sprintf(strings.Join(result, " "))
}
func (a *Attributes) MarshalJSON() ([]byte, error) {
	result := make(map[string]string)
	for _, v := range a.Attributes() {
		result[v.Key] = v.Value
	}
	return json.Marshal(result)
}

func (a *Attributes) SetAttribute(attr Attribute) error {
	if attr.Key == "" {
		return nil
	}
	for i, v := range *a {
		if v.Key == attr.Key {
			if attr.Value == "" {
				(*a)[i] = (*a)[len(*a)-1]
				*a = (*a)[:len(*a)-1]
				return nil
			}
			(*a)[i].Value = attr.Value
			return nil
		}
	}
	if attr.Value != "" {
		*a = append(*a, attr)
	}
	return nil
}
