package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

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

func main() {

	port := os.Getenv("PORT")

	if port == "" {
		port = "3000"
	}

	app := NewApp()

	app.Handle(`^/hash$`, func(ctx *Context) {
		log.Print("Post hash request received")
		var startTime = time.Now()

		if ctx.Request.Method == http.MethodPost {
			wg.Add(1)

			var newValue = atomic.AddUint64(&hashIdCounter, 1)
			ctx.Request.ParseForm()
			var password = ctx.Request.PostFormValue("password")

			var timer = time.NewTimer(5 * time.Second)
			go func() {
				<-timer.C
				hashedPasswords.Store(newValue, GetHashedPassword(password))
				wg.Done()
			}()

			ctx.Text(http.StatusOK, fmt.Sprintf("%v", newValue))

			var duration = time.Now().Sub(startTime)
			var processingTime = uint64((duration * time.Nanosecond))
			log.Printf("Post hash request processed in %vns", processingTime)

			mutex.Lock()
			stats.Average = ((stats.Average * (hashIdCounter - 1)) + processingTime/1000) / hashIdCounter
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
