package main

import (
	"net/http"
	"regexp"
)

type Handler func(*Context)

type Route struct {
	Pattern *regexp.Regexp
	Handler Handler
}

type App struct {
	Routes       []Route
	DefaultRoute Handler
}

func NewApp() *App {
	app := &App{
		DefaultRoute: func(ctx *Context) {
			ctx.Text(http.StatusNotFound, "Not found")
		},
	}

	return app
}

func (a *App) Handle(pattern string, handler Handler) {
	re := regexp.MustCompile(pattern)
	route := Route{Pattern: re, Handler: handler}

	a.Routes = append(a.Routes, route)
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := &Context{Request: r, ResponseWriter: w}

	if isShuttingDown {
		ctx.Text(http.StatusServiceUnavailable, "Service Unavailable")
	} else {
		ctx := &Context{Request: r, ResponseWriter: w}

		for _, rt := range a.Routes {
			if matches := rt.Pattern.FindStringSubmatch(ctx.URL.Path); len(matches) > 0 {
				if len(matches) > 1 {
					ctx.Params = matches[1:]
				}

				rt.Handler(ctx)

				return
			}
		}

		a.DefaultRoute(ctx)
	}
}
