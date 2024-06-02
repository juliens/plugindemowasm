// Package plugindemo a demo plugin.
package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"text/template"

	"github.com/http-wasm/http-wasm-guest-tinygo/handler"
	"github.com/http-wasm/http-wasm-guest-tinygo/handler/api"
	"github.com/juliens/wasm-goexport/guest"
	_ "github.com/stealthrocket/net/http"
)

func main() {
	var config Config
	err := json.Unmarshal(handler.Host.GetConfig(), &config)
	if err != nil {
		handler.Host.Log(api.LogLevelError, fmt.Sprintf("Could not load config %v", err))
		os.Exit(1)
	}

	mw, err := New(config)
	if err != nil {
		handler.Host.Log(api.LogLevelError, fmt.Sprintf("Could not load config %v", err))
		os.Exit(1)
	}
	handler.HandleRequestFn = mw.handleRequest

	guest.SetExports(handler.GetExports())
}

// Config the plugin configuration.
type Config struct {
	Headers map[string]string `json:"headers,omitempty"`
	Remote  string            `json:"remote"`
}

// Demo a Demo plugin.
type Demo struct {
	headers  map[string]string
	remote   string
	template *template.Template
}

// New created a new Demo plugin.
func New(config Config) (*Demo, error) {
	if len(config.Headers) == 0 {
		return nil, fmt.Errorf("headers cannot be empty")
	}

	return &Demo{
		headers:  config.Headers,
		remote:   config.Remote,
		template: template.New("demo").Delims("[[", "]]"),
	}, nil
}

func (a *Demo) handleRequest(req api.Request, resp api.Response) (next bool, reqCtx uint32) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	rp, err := http.Get(a.remote)
	if err != nil {
		resp.SetStatusCode(http.StatusInternalServerError)
		resp.Body().Write([]byte(err.Error()))
		return false, 0
	}

	req.Headers().Set("Remote-Status", rp.Status)

	body, err := io.ReadAll(rp.Body)
	if err != nil {
		resp.SetStatusCode(http.StatusInternalServerError)
		resp.Body().Write([]byte(err.Error()))
		return false, 0
	}

	if err != nil {
		resp.SetStatusCode(http.StatusInternalServerError)
		resp.Body().Write([]byte(err.Error()))
		return false, 0
	}

	re, err := regexp.Compile("<title[^>]*>([^<]*)</title>")
	if err != nil {
		log.Fatal(err)
	}
	all := re.FindAllSubmatch(body, -1)
	if len(all) > 0 {
		if len(all[0]) > 1 {
			req.Headers().Set("Remote-Title", string(all[0][1]))
		}
	}

	if _, ok := req.Headers().Get("debug"); ok {
		fmt.Println("body", string(body))
	}
	req.Headers().Set("Remote-Body-Begin", string(body[0:10]))

	var toto map[string]interface{}
	err = json.Unmarshal([]byte(`{"Test":"Value"}`), &toto)
	if err != nil {
		resp.SetStatusCode(http.StatusInternalServerError)
		resp.Body().Write([]byte(err.Error()))
		return false, 0
	}

	req.Headers().Set("Json-Value", toto["Test"].(string))

	for key, value := range a.headers {
		tmpl, err := a.template.Parse(value)
		if err != nil {
			resp.SetStatusCode(http.StatusInternalServerError)
			resp.Body().Write([]byte(err.Error()))
			return false, 0
		}

		writer := &bytes.Buffer{}

		err = tmpl.Execute(writer, req)
		if err != nil {
			resp.SetStatusCode(http.StatusInternalServerError)
			resp.Body().Write([]byte(err.Error()))
			return false, 0
		}

		req.Headers().Set(key, writer.String())
	}

	return true, 0
}
