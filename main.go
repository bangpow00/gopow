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

var elapsedChan = make(chan time.Duration, 100)
var passwordChan = make(chan passwordJob, 100)
var elapsedUsecs int64
var counter int64

type passwordJob struct {
	id     int64
	passwd string
}

func runTime(start time.Time) {
	elapsed := time.Since(start)
	select {
	case elapsedChan <- elapsed:
		// dont block
	default:
		fmt.Println("elapsedChan write would block")
	}
}

func storePasswordWorker(job passwordJob, wg *sync.WaitGroup) {
	defer wg.Done()

	start := time.Now()
	passwdHash := hasher.EncodeSha512Base64(job.passwd)
	time.Sleep((5 * time.Second) - time.Since(start))
	safemapper.Add(job.id, passwdHash)
}

func storePasswordJobScheduler(jobs <-chan passwordJob, wg *sync.WaitGroup) {
	for job := range jobs {
		wg.Add(1)
		go storePasswordWorker(job, wg)
	}
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

		passwordChan <- passwordJob{id: id, passwd: passwd}
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

	var wg sync.WaitGroup
	go storePasswordJobScheduler(passwordChan, &wg)

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
	// to ignore 0s. A few correct elasped times are better than a bunch of wrong
	// ones.
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
