package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/bangpow00/gopow/handlers"
	"github.com/bangpow00/gopow/hasher"
	"github.com/bangpow00/gopow/safemapper"
)

func storePasswordWorker(id int64, passwd string, wg *sync.WaitGroup) {
	defer wg.Done()

	start := time.Now()
	passwdHash := hasher.EncodeSha512Base64(passwd)
	time.Sleep((5 * time.Second) - time.Since(start))
	safemapper.Add(id, passwdHash)
}

func storePasswordJobScheduler(jobs <-chan map[int64]string, wg *sync.WaitGroup) {
	for job := range jobs {
		for key, value := range job {
			wg.Add(1)
			go storePasswordWorker(key, value, wg)
		}
	}
}

func main() {

	passwdjobs := make(chan map[int64]string, 100)

	handler := http.NewServeMux()

	handler.HandleFunc("/hash/", handlers.GetPasswordHandler)
	handler.HandleFunc("/hash", handlers.CreatePasswordHandler(passwdjobs))
	handler.HandleFunc("/stats", handlers.StatsHandler)

	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	})

	var wg sync.WaitGroup
	go storePasswordJobScheduler(passwdjobs, &wg)

	doShutdown := make(chan os.Signal)
	signal.Notify(doShutdown, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-doShutdown
		fmt.Println("Caught", sig, "but waiting for threads to complete")
		wg.Wait()
		os.Exit(0)
	}()

	log.Fatal(http.ListenAndServe(":8080", handler))
}
