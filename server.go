package main

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
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

type Stats struct {
	Total   uint64
	Average uint64
}

var stats = Stats{Total: 0, Average: 0}
var hashedPasswords = new(sync.Map)
var hashIdCounter uint64

var wg sync.WaitGroup
var shutdownWg sync.WaitGroup
var mutex = &sync.Mutex{}
var numberOfRequests uint64
var isShuttingDown bool

func GetHashedPassword(password string) string {

	sha_512 := sha512.New()
	sha_512.Write([]byte(password))

	return base64.StdEncoding.EncodeToString(sha_512.Sum(nil))
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

func main() {

	port := os.Getenv("PORT")

	if port == "" {
		port = "3000"
	}

	app := NewApp()

	app.Handle(`^/hash$`, func(ctx *Context) {
		startTime := time.Now()

		if ctx.Request.Method == http.MethodPost {
			wg.Add(1)

			var newValue = atomic.AddUint64(&hashIdCounter, 1)
			ctx.Request.ParseForm()
			var password = ctx.Request.PostFormValue("password")

			var timer = time.NewTimer(30 * time.Second)
			go func() {
				<-timer.C
				hashedPasswords.Store(newValue, GetHashedPassword(password))
				wg.Done()
			}()

			ctx.Text(http.StatusOK, fmt.Sprintf("%v", newValue))

			duration := time.Now().Sub(startTime)

			mutex.Lock()
			var processingTime = uint64((duration * time.Microsecond))
			stats.Average = ((stats.Average * (hashIdCounter - 1)) + processingTime) / hashIdCounter
			stats.Total = hashIdCounter
			mutex.Unlock()

		} else {
			ctx.Text(http.StatusNotFound, "Invalid Method")
		}

	})

	app.Handle(`/hash/([0-9]+)$`, func(ctx *Context) {
		if ctx.Request.Method == http.MethodGet {
			var hashId = ctx.Params[0]
			var hashId64, _ = strconv.ParseUint(hashId, 10, 64)

			var hashedPassword, ok = hashedPasswords.Load(hashId64)
			if !ok {
				// handle error
				ctx.Text(http.StatusNotFound, "Not Found")

			} else {

				ctx.Text(http.StatusOK, fmt.Sprintf("%v", hashedPassword))
			}
		} else {
			ctx.Text(http.StatusNotFound, "Invalid Method")
		}
	})

	app.Handle(`/stats`, func(ctx *Context) {
		ctx.Json(http.StatusOK, stats)
	})

	app.Handle(`/shutdown`, func(ctx *Context) {
		ctx.Text(http.StatusOK, "")
		shutdownWg.Done()
		isShuttingDown = true
	})

	shutdownWg.Add(1)

	go func() {
		err := http.ListenAndServe(":"+port, app)

		if err != nil {
			log.Fatalf("Could not start server: %s\n", err.Error())
		}
	}()

	shutdownWg.Wait()

	log.Printf("Server shutdown: started")

	wg.Wait()

	log.Printf("Server shutdown: finished")

}
