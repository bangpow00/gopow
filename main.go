package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/bangpow00/gopow/hasher"
	"github.com/bangpow00/gopow/safecounter"
	"github.com/bangpow00/gopow/safeelapsed"
	"github.com/bangpow00/gopow/safemapper"
)

var wg sync.WaitGroup

type elapsedTime struct {
	elapsed time.Duration
}

var elapsedTimeQueue = make(chan elapsedTime, 100)

func runTime(start time.Time) {
	safeelapsed.Add(time.Since(start))
}

func calculateAverageTime() {

}

func storePasswordHash(transID int, passwd string) {
	defer wg.Done()

	start := time.Now()
	passwdHash := hasher.EncodeSha512Base64(passwd)
	time.Sleep((5 * time.Second) - time.Since(start))
	safemapper.Add(transID, passwdHash)
}

func createPasswordHashHandler(w http.ResponseWriter, r *http.Request) {
	defer runTime(time.Now())

	switch r.Method {
	case http.MethodPost:
		err := r.ParseForm()
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		passwd := r.Form.Get("password")
		if passwd == "" {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		id := safecounter.GetUnique()
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, fmt.Sprintf("%d", id))

		// Bangpow: think about a race condition here which could result in a waitgroup
		// hang
		wg.Add(1)
		go storePasswordHash(id, passwd)

	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func getPasswordHashHandler(w http.ResponseWriter, r *http.Request) {
	defer runTime(time.Now())

	switch r.Method {
	case http.MethodGet:
		idStr := strings.Replace(r.URL.Path, "/hash/", "", 1)
		id, _ := strconv.Atoi(idStr)
		hashval := safemapper.Value(id)
		if hashval == "" {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, fmt.Sprintf("%s", hashval))
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}

}

func statsHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodGet:
		m := map[string]int{"total": safemapper.Len(), "average": safeelapsed.Average()}
		jsonBody, _ := json.Marshal(m)
		w.WriteHeader(http.StatusOK)
		w.Write(jsonBody)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func main() {

	handler := http.NewServeMux()

	handler.HandleFunc("/hash/", getPasswordHashHandler)
	handler.HandleFunc("/hash", createPasswordHashHandler)
	handler.HandleFunc("/stats", statsHandler)

	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	})

	doShutdown := make(chan os.Signal)
	signal.Notify(doShutdown, syscall.SIGTERM)
	signal.Notify(doShutdown, syscall.SIGINT)
	go func() {
		sig := <-doShutdown
		fmt.Printf("Caught \"%+v\" but waiting for threads to complete\n", sig)
		wg.Wait()
		os.Exit(0)
	}()

	log.Fatal(http.ListenAndServe(":8080", handler))
}
