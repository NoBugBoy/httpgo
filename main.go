package main

import (
	"fmt"
	"github.com/NoBugBoy/httpgo/test"
	_ "github.com/NoBugBoy/httpgo/test"
	"runtime"
	"time"
)

func main() {
	start := time.Now()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("%+v\n", m.TotalAlloc)
	test.Test7()
	runtime.ReadMemStats(&m)
	fmt.Printf("%+v\n", m.TotalAlloc)
	end := time.Since(start)
	fmt.Println("time", end)
}
