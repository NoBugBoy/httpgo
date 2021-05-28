package main

import (
	"fmt"
	"github.com/NoBugBoy/httpgo/test"
	"time"
)

func main() {
	start := time.Now()
	test.Test5()
	fmt.Println(time.Now().Second() - start.Second())
}
