package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bangpow00/gopow/Hasher"
)

type ElapasedTime struct {
	start int
	end   int
}

type SafeCounter struct {
	count int
	mux   sync.Mutex
}

type SafeHash struct {
	hashmap map[int]string
	mux     sync.Mutex
}

var transactionId SafeCounter
var safeHash SafeHash

func (c *SafeCounter) getNext() int {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.count++
	return c.count
}

func (c *SafeCounter) getCurrent() int {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.count
}

func (h *SafeHash) get(transID int) string {
	h.mux.Lock()
	defer h.mux.Unlock()
	return h.hashmap[transID]
}

func (h *SafeHash) set(transID int, hsh string) {
	h.mux.Lock()
	defer h.mux.Unlock()
	h.hashmap[transID] = hsh
}

func storePasswordHash(transID int, passwd string) {
	// including encode time for the 5 second wait
	start := time.Now()
	passwdHash := Hasher.EncodeSha512Base64(passwd)
	time.Sleep((5 * time.Second) - time.Since(start))
	safeHash.set(transID, passwdHash)
}

func createHashHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			panic(err)
		}

		passwd := r.Form.Get("password")
		if passwd != "" {
			id := transactionId.getNext()
			go storePasswordHash(id, passwd)

			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)

			io.WriteString(w, fmt.Sprintf("%d", id))
		} else {
			http.Error(w, "Invalid Parameters", http.StatusBadRequest)
		}
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

func returnHashHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		idStr := strings.Replace(r.URL.Path, "/hash/", "", 1)
		id, _ := strconv.Atoi(idStr)
		hash := safeHash.get(id)
		if hash == "" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, fmt.Sprintf("%s", hash))
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}

}

func statsHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodGet:
		mapFoo := map[string]int{"total": transactionId.count, "average": 123}
		jsonBody, _ := json.Marshal(mapFoo)
		w.WriteHeader(http.StatusOK)
		w.Write(jsonBody)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func main() {
	safeHash = SafeHash{hashmap: make(map[int]string)}

	handler := http.NewServeMux()

	handler.HandleFunc("/hash/", returnHashHandler)
	handler.HandleFunc("/hash", createHashHandler)
	handler.HandleFunc("/stats", statsHandler)

	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	})

	log.Fatal(http.ListenAndServe(":8080", handler))
}
