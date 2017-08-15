package jsonflag

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func TestLiteralInput(t *testing.T) {
	var flg Var

	// Good JSON
	if err := flg.Set(`{"foo": "bar"}`); err != nil {
		t.Fatalf("Execpted set to succeed but got: %v", err)
	}
	expected := map[string]interface{}{
		"foo": "bar",
	}
	if flg.Value == nil || !reflect.DeepEqual(*flg.Value, expected) {
		t.Fatalf("Expected set value to be %v but got: %v", expected, flg.Value)
	}

	// Bad JSON
	if err := flg.Set(`{"foo": "bar`); err == nil {
		t.Fatalf("Expected set to fail")
	}

	// UseNumber
	flg.UseNumber = true
	if err := flg.Set(`{"foo": 1}`); err != nil {
		t.Fatalf("Execpted set to succeed but got: %v", err)
	}
	expected = map[string]interface{}{
		"foo": json.Number("1"),
	}
	if flg.Value == nil || !reflect.DeepEqual(*flg.Value, expected) {
		t.Fatalf("Expected set value to be %v but got: %v", expected, flg.Value)
	}

}

func TestFileInput(t *testing.T) {
	var flg Var
	f, err := ioutil.TempFile("", "jsonflag_test")
	if err != nil {
		panic(err)
	}
	defer os.Remove(f.Name())

	// Good JSON
	if _, err := f.Write([]byte(`{"foo": "bar"}`)); err != nil {
		panic(err)
	}

	if err := flg.Set("@" + f.Name()); err != nil {
		t.Fatalf("Expected set to succeed but got: %v", err)
	}

	expected := map[string]interface{}{
		"foo": "bar",
	}
	if flg.Value == nil || !reflect.DeepEqual(*flg.Value, expected) {
		t.Fatalf("Expected set value to be %v but got: %v", expected, flg.Value)
	}

	// Bad JSON
	if err := f.Truncate(0); err != nil {
		panic(err)
	}

	if _, err := f.Write([]byte(`{"foo": "bar`)); err != nil {
		panic(err)
	}

	if err := flg.Set("@" + f.Name()); err == nil {
		t.Fatalf("Expected set to fail")
	}

	// Bad file
	if err := flg.Set("@" + f.Name() + "non-existent-file"); err == nil {
		t.Fatalf("Expected set to fail")
	}
}

func TestUndefined(t *testing.T) {
	var flg Var
	if err := flg.Set(""); err != nil {
		t.Fatalf("Expected set to succeed but got: %v", err)
	}
	if flg.Value != nil {
		t.Fatalf("Expected undefined")
	}
}
