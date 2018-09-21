package safeelapsed

import (
	"sync"
	"time"
)

var totalelapsed time.Duration
var count int
var mux sync.Mutex

// Add up times
func Add(elapsed time.Duration) {
	//mux.Lock()
	//defer mux.Unlock()
	//d := int64(elapsed) / time.Microsecond
	//fmt.Println("total: ", totalelapsed, "elapsed: ", d)
	//totalelapsed += d
	//count++
}

//Average time
func Average() int {
	//mux.Lock()
	//defer mux.Unlock()
	//fmt.Println("total: ", totalelapsed, "Count: ", count)
	//ave := int(totalelapsed) / count
	return 123
}
