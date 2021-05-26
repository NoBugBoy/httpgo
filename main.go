package main

import (
	"fmt"
	. "http-go/http"
	"net/http"
)

func main() {
		req := &Req{}
		str := req.Url("http://127.0.0.1:8080/put").
			Method(http.MethodPut).
			Header("Content-Type","application/json").
			Params(Query{
				"id":11,
				"aaa": "123123",
			}).
			Go().
			Body()
		fmt.Println(str)
}
