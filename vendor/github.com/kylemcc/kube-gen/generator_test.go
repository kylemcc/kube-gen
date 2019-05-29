package kubegen

import (
	"errors"
	"reflect"
	"testing"
)

func TestValidateConfig(t *testing.T) {
	cases := []struct {
		input    *generator
		expected error
	}{
		{&generator{}, nil},
		{&generator{Config: Config{ResourceTypes: []string{"pods", "services", "endpoints"}}}, nil},
		{&generator{Config: Config{ResourceTypes: []string{"invalidtype", "services", "endpoints"}}}, errors.New("invalid type: invalidtype")},
	}

	for i, c := range cases {
		if err := c.input.validateConfig(); !reflect.DeepEqual(err, c.expected) {
			t.Errorf("case %d failed: got [%#v] expected [%#v]\n", i, err, c.expected)
		}
	}
}
