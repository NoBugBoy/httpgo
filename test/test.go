package main

import (
	"fmt"
	. "http-go/http"
	"net/http"
	"sync"
)
// 压力测试
var join sync.WaitGroup
func main() {
	arr := make([]*Req,1000)

	for i := 0 ; i < 1000 ; i++ {
		join.Add(1)
		req := &Req{}
		x := req.Url("http://www.youtube.com").
			Method(http.MethodGet).
			Header("Content-Type","application/json").
			Params(Query{
				"id":"1",
			}).
			Timeout(1).
			Build()
		arr[i] = x
	}
	for _, req := range arr {
		go runAndPrint(req)
	}
	join.Wait()
}
func runAndPrint(r *Req)  {
	defer join.Done()
	fmt.Println(r.Go().Body())
}
