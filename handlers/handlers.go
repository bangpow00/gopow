package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/bangpow00/gopow/safemapper"
)

var counter int64
var elapsedUsecs int64
var elapsedChan = make(chan time.Duration, 100)

func init() {
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

//CreatePasswordHandler wrapper
func CreatePasswordHandler(jobs chan<- map[int64]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

			jobs <- map[int64]string{id: passwd}
		default:
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
	}
}

// GetPasswordHandler wrapper
func GetPasswordHandler(w http.ResponseWriter, r *http.Request) {

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

// StatsHandler for returning server stats
func StatsHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodGet:
		m := map[string]int64{"total": atomic.LoadInt64(&counter), "average": atomic.LoadInt64(&elapsedUsecs)}
		stats, _ := json.Marshal(m)
		w.WriteHeader(http.StatusOK)
		w.Write(stats)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}
