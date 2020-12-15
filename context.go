package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Context struct {
	http.ResponseWriter
	*http.Request
	Params []string
}

func (c *Context) Text(code int, body string) {
	c.ResponseWriter.Header().Set("Content-Type", "text/plain")
	c.WriteHeader(code)

	io.WriteString(c.ResponseWriter, fmt.Sprintf("%s\n", body))
}

func (c *Context) Json(code int, body interface{}) {
	c.ResponseWriter.Header().Set("Content-Type", "application/json")
	c.WriteHeader(code)
	var en = json.NewEncoder(c.ResponseWriter)
	en.SetEscapeHTML(false)
	en.Encode(body)
}
