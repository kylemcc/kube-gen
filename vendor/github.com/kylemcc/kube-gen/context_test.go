package kubegen

import (
	"os"
	"reflect"
	"testing"
)

func TestContextEnv(t *testing.T) {
	os.Clearenv()
	os.Setenv("SOMEKEY", "some value")
	os.Setenv("KEY_WITH_EMPTY_VALUE", "")
	os.Setenv("HAS_MULTIPLE_EQ", "value=foo=baz")

	expected := map[string]string{
		"SOMEKEY":              "some value",
		"KEY_WITH_EMPTY_VALUE": "",
		"HAS_MULTIPLE_EQ":      "value=foo=baz",
	}

	ctx := Context{}
	if !reflect.DeepEqual(ctx.Env(), expected) {
		t.Errorf("environment did not parse correctly. Expected [%#v] got [%#v]\n", expected, ctx.Env())
	}
}

func TestContextEnvParsedOnce(t *testing.T) {
	os.Clearenv()
	os.Setenv("SOMEKEY", "some value")
	os.Setenv("KEY_WITH_EMPTY_VALUE", "")
	os.Setenv("HAS_MULTIPLE_EQ", "value=foo=baz")

	ctx := Context{}
	first := ctx.Env()

	os.Clearenv()
	os.Setenv("TOTALLY_DIFFERENT", "value")

	second := ctx.Env()

	if !reflect.DeepEqual(first, second) {
		t.Errorf("Context.Env should only parse the environment once. Expected [%#v] on second call. Got [%#v]\n", first, second)
	}
}
