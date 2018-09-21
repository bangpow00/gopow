package safemapper

import "sync"

// SafeMap is thread safe map
//type SafeMap struct {
//	m   map[int]string
//	mux sync.Mutex
//}

var m map[int64]string
var mux sync.Mutex

func init() {
	m = make(map[int64]string)
}

// Value return value for a key
func Value(key int64) string {
	mux.Lock()
	defer mux.Unlock()
	return m[key]
}

// Add key and value pair to map
func Add(key int64, val string) {
	mux.Lock()
	defer mux.Unlock()
	m[key] = val
}
