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
	"sync/atomic"
	"syscall"
	"time"

	"github.com/bangpow00/gopow/hasher"
	"github.com/bangpow00/gopow/safemapper"
)

var wg sync.WaitGroup
var elapsedChan = make(chan time.Duration, 100)
var elapsedUsecs int64
var counter int64

func runTime(start time.Time) {
	elapsed := time.Since(start)
	select {
	case elapsedChan <- elapsed:
		// dont block
	default:
		fmt.Println("elapsedChan write would block")
	}
}

func storePasswordHash(transID int64, passwd string) {
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

		id := atomic.AddInt64(&counter, 1)
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, fmt.Sprintf("%d", id))

		wg.Add(1)
		go storePasswordHash(id, passwd)

	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func getPasswordHashHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodGet:
		idStr := strings.Replace(r.URL.Path, "/hash/", "", 1)
		id, _ := strconv.ParseInt(idStr, 10, 64)
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
		cnt := atomic.LoadInt64(&counter)
		m := map[string]int64{"total": cnt, "average": atomic.LoadInt64(&elapsedUsecs)}
		stats, _ := json.Marshal(m)
		w.WriteHeader(http.StatusOK)
		w.Write(stats)
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
	signal.Notify(doShutdown, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-doShutdown
		fmt.Println("Caught", sig, "but waiting for threads to complete")
		wg.Wait()
		os.Exit(0)
	}()

	// on my system I'm seeing an elapsed time of 0 most of the time.
	// I've puzzled over this for quite a while. Tried sprinkling in time.Now()
	// at the start and end of the handler function and am seeing the same value.
	// Maybe it's my bug, but I'm not seeing it. So my phase 1 workaround is
	// to ignore 0s.
	go func() {
		var totalelapsed int64
		var cnt int64
		for elapsed := range elapsedChan {
			if elapsed != 0 {
				cnt++
				totalelapsed += int64(elapsed)
				x := totalelapsed / cnt
				atomic.StoreInt64(&elapsedUsecs, x/1000)
			}
		}
	}()

	log.Fatal(http.ListenAndServe(":8080", handler))
}
