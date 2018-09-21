package safeelapsed

import (
	"sync"
	"time"
)

var totalelapsedms int64
var count int64
var mux sync.Mutex
var times []int64

//func init() {
//	times = make([]int64, 1, 10000)
//}

// Add up times
func Add(elapsed time.Duration) {
	mux.Lock()
	defer mux.Unlock()
	count++
	totalelapsedms += int64(elapsed / time.Microsecond)
}

// Average without slice
func Average() int {
	mux.Lock()
	defer mux.Unlock()
	ave := totalelapsedms / count
	//str := strconv.FormatInt(ave, 10)
	return int(ave)
}
