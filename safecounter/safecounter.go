package safecounter

import (
	"sync"
)

var count int
var mux sync.Mutex

//GetUnique increment the counter
func GetUnique() int {
	mux.Lock()
	defer mux.Unlock()
	count++
	return count
}
