package safemapper

import "sync"

// SafeMap is thread safe map
//type SafeMap struct {
//	m   map[int]string
//	mux sync.Mutex
//}

var m map[int]string
var mux sync.Mutex

func init() {
	m = make(map[int]string)
}

// Value return value for a key
func Value(key int) string {
	mux.Lock()
	defer mux.Unlock()
	return m[key]
}

// Add key and value pair to map
func Add(key int, val string) {
	mux.Lock()
	defer mux.Unlock()
	m[key] = val
}

// Len return length of map
func Len() int {
	// POWBUG: I am concerned about the performce bottleneck of locking
	// on len() when the map is large.
	mux.Lock()
	defer mux.Unlock()
	return len(m)
}
