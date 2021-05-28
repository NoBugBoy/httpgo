package main

import (
	"fmt"
	"github.com/NoBugBoy/httpgo/test"
	"time"
)

func main() {
	start := time.Now()
	test.Test4()
	fmt.Println(time.Now().Second() - start.Second())
}
