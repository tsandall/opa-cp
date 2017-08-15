// Copyright 2017 Torin Sandall.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package jsonflag

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"strings"
)

// Var implements the github.com/spf13/pflag#Value interface to load JSON
// encoded values from command line arguments. E.g.,
//
//	var flg jsonflag.Var
//	...
//	cmd.Flags().VarP(&flg, "my-flag", "", "set my flag")
//	...
//  if flg.Value != nil { // is set }
//
// And then the flag can be used as follows:
//
//	$ cmd --my-flag @filename.json
//	$ cmd --my-flag '{"foo": "bar"}'
type Var struct {
	UseNumber bool         // Use json.Number for numerics
	Value     *interface{} // Loaded value
}

// Set sets the flag from the string s.
func (flg *Var) Set(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	var bs []byte
	if strings.HasPrefix(s, "@") {
		var err error
		bs, err = ioutil.ReadFile(s[1:])
		if err != nil {
			return err
		}
	} else {
		bs = []byte(s)
	}
	buf := bytes.NewBuffer(bs)
	decoder := json.NewDecoder(buf)
	if flg.UseNumber {
		decoder.UseNumber()
	}
	var v interface{}
	if err := decoder.Decode(&v); err != nil {
		return err
	}
	flg.Value = &v
	return nil
}

// Type returns a string indicating how to supply the flag's value.
func (flg Var) Type() string {
	return "@filename or JSON literal"
}

func (flg Var) String() string {
	if flg.Value == nil {
		return "undefined"
	}
	bs, err := json.Marshal(*flg.Value)
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	if err := json.Compact(&buf, bs); err != nil {
		panic(err)
	}
	return string(buf.Bytes()[:10])
}
