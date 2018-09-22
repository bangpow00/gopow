package main

import (
	"fmt"
	"math/big"
	"testing"
	"time"
)

func TestTime(t *testing.T) {
	var errs int
	for i := 0; i < 10; i++ {
		start := time.Now()

		r := new(big.Int)
		fmt.Println(r.Binomial(1000, 10))

		if time.Since(start) == 0 {
			errs++
		}
	}
	if errs != 0 {
		t.Errorf("Yup, got elapsed=0 [%d] times", errs)
	}
}
