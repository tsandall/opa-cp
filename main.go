// Copyright 2017 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/tsandall/jsonflag"
)

const (
	defaultOPAURL = "http://localhost:8181"
)

type config struct {
	url     string
	headers map[string]string
	output  string
	input   *interface{}
	delay   time.Duration
}

func main() {

	cfg := config{
		headers: map[string]string{
			"Authorization": os.Getenv("AUTH"),
		},
	}

	var baseURL string
	var inputFlag jsonflag.Var

	cmd := &cobra.Command{
		Use: "opa-cp path [path]",
		Long: `CLI tool to copy OPA documents.

Copy data.example.path to stdout using input.json as input document:

	$ opa-cp /example/path --input @input.json

Copy data.example.path to ./local/directory/file.json:

	$ opa-cp /example/path ./local/directory/file.json

Copy data.example.path to stdout and include HTTP Authorization header in request:

	$ AUTH="Bearer secret-token" opa-cp /example/path
`,
		Run: func(cmd *cobra.Command, args []string) {
			cfg.input = inputFlag.Value
			path := strings.Trim(args[0], "/")
			if path != "" {
				path = "/" + path
			}
			cfg.url = strings.TrimRight(baseURL, "/") + "/v1/data" + path
			if len(args) > 1 {
				cfg.output = args[1]
			}
			run(cfg)
		},
		Args: cobra.RangeArgs(1, 2),
	}

	opaURL := os.Getenv("OPA_URL")
	if opaURL == "" {
		opaURL = defaultOPAURL
	}

	cmd.Flags().StringVarP(&baseURL, "url", "u", opaURL, "set OPA root URL")
	cmd.Flags().DurationVarP(&cfg.delay, "delay", "d", time.Minute*1, "set polling delay")
	cmd.Flags().VarP(&inputFlag, "input", "i", "set input document")

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(cfg config) {
	for {
		if err := oneShot(cfg); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		time.Sleep(cfg.delay)
	}
}

func oneShot(cfg config) error {

	c := &client{
		url:     cfg.url,
		headers: cfg.headers,
		input:   cfg.input,
	}

	result, err := c.Do()
	if err != nil {
		return errors.Wrapf(err, "request failed")
	}

	if result == nil {
		return nil
	}

	var bs []byte

	switch r := (*result).(type) {
	case string:
		bs = []byte(r)
	default:
		return fmt.Errorf("bad result type %T: expected string", *result)
	}

	if cfg.output != "" {
		if err := os.MkdirAll(path.Dir(cfg.output), 755); err != nil {
			return errors.Wrapf(err, "mkdir failed")
		}
		if err := ioutil.WriteFile(cfg.output, bs, 0644); err != nil {
			return errors.Wrapf(err, "write failed")
		}
	} else {
		fmt.Println(string(bs))
	}

	return nil
}

type client struct {
	url     string
	input   *interface{}
	headers map[string]string
}

type dataRequestV1 struct {
	Input *interface{} `json:"input,omitempty"`
}

type dataResponseV1 struct {
	Result *interface{} `json:"result,omitempty"`
}

type errorResponseV1 struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e errorResponseV1) Error() string {
	return fmt.Sprintf("%v: %v", e.Code, e.Message)
}

func (c *client) Do() (*interface{}, error) {

	var buf bytes.Buffer
	method := "GET"

	if c.input != nil {
		body := dataRequestV1{
			Input: c.input,
		}
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return nil, err
		}
		method = "POST"
	}

	req, err := http.NewRequest(method, c.url, &buf)
	if err != nil {
		return nil, err
	}

	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, c.handleError(resp)
	}

	return c.handleSuccess(resp)
}

func (c *client) handleError(resp *http.Response) error {
	if strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
		var err errorResponseV1
		if err := json.NewDecoder(resp.Body).Decode(&err); err != nil {
			return err
		}
		return err
	}
	return fmt.Errorf("status %v: unknown error", resp.StatusCode)
}

func (c *client) handleSuccess(resp *http.Response) (*interface{}, error) {
	var result dataResponseV1
	return result.Result, json.NewDecoder(resp.Body).Decode(&result)
}
