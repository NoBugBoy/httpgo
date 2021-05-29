package http

import "net/http"

// QuickSend 快速发送指定请求
func QuickSend() func(url string,method string,param Query) (string,error) {
	r := &Req{}
	return func(url string,method string,param Query) (string,error){
		return r.Method(method).Url(url).Params(param).Build().Go().Body()
	}
}

// QuickSend0 快速发送get请求
func QuickSend0() func(url string) (string,error) {
	r := &Req{}
	return func(url string) (string,error){
		return r.Url(url).Method(http.MethodGet).Build().Go().Body()
	}
}
