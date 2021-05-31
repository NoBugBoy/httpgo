package test

import (
	"fmt"
	. "github.com/NoBugBoy/httpgo/http"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
	"time"
)

// Test7 测试QuickSend0
func Test7()  {
	//get := QuickSend0()
	//_, _ = get("http://localhost:8080/get/1")
	req := &Req{}
	_, _ = req.
		Method(http.MethodPost).
		Url("http://localhost:8080/post").
		Header("Content-Type", "application/json").
		Params(Query{
			"id": "123",
			"aaa" :"1123",
		}).
		Timeout(30).
		Go(). //只有调用Go才会发起请求，并且在该方法内进行连接关闭防止泄露
		Body()
}

// Test6 测试时间和内存
func Test6()  {
	start := time.Now()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("%+v\n", m.TotalAlloc)
	Test1()
	runtime.ReadMemStats(&m)
	fmt.Printf("%+v\n", m.TotalAlloc)
	end := time.Since(start)
	fmt.Println("time", end)
}

// Test4 测试response close
func Test4() {
	req := &Req{}
	_, _ = req.ImportProxy().
		Method(http.MethodGet).
		Url("http://www.baidu.com?a=1").
		Params(Query{
			"x": "123",
		}).
		Timeout(30).
		Go(). //只有调用Go才会发起请求，并且在该方法内进行连接关闭防止泄露
		Body()
	close, _ := ioutil.ReadAll(req.Response.Body)
	if len(close) == 0 {
		fmt.Println("连接已经关闭")
	} else {
		fmt.Println("连接未关闭")
	}

}
func startChunkServer() {
	http.HandleFunc("/get",
		func(writer http.ResponseWriter, request *http.Request) {
			flusher, ok := writer.(http.Flusher)
			if !ok {
				panic("type err")
			}
			for i := 0; i < 10; i++ {
				fmt.Fprintf(writer, "data [%d] %d \n", i, rand.Intn(999)+rand.Int())
				flusher.Flush()
				time.Sleep(time.Second * 2)
			}
		})
	http.ListenAndServe(":8080", nil)
}

// Test5 测试分块传输
func Test5() {
	go startChunkServer()
	req := &Req{}
	re := req.ImportProxy().
		Method(http.MethodGet).
		Header("Connection", "Keep-Alive").
		Header("Transfer-Encoding", "chunked").
		Url("http://localhost:8080/get").
		Chunk().
		Timeout(30). //超时会关闭
		Go()
	fmt.Println(re.Response.Header)
	data := make([]byte, 1024)
	for {
		read, err := re.Response.Body.Read(data)
		fmt.Println("字节长度 ", read)
		if read > 0 {
			fmt.Print(string(data[:read]))
		}
		if err == io.EOF {
			break
		}
	}
	fmt.Println("Ok")
}

// Test3 测试Params和pathQuery同时存在应该优先使用pathQuery
func Test3() {
	req := &Req{}
	result, _ := req.ImportProxy().
		Method(http.MethodGet).
		Url("http://www.baidu.com?a=1").
		Params(Query{
			"x": "123",
		}).
		Timeout(30).Go().Body()
	fmt.Println(result)
}

// Test2 代理测试 关注返回关键字 "您的IP"
func Test2() {
	req := &Req{}
	result, _ := req.ImportProxy().
		Method(http.MethodGet).
		Header("Connection", "Keep-Alive").
		Header("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36").
		Url("http://ip.tool.chinaz.com/").
		Proxy().
		Timeout(30).
		Go().
		Body()
	fmt.Println(result)
}

var join sync.WaitGroup

// Test1 压力测试, 注意 ulimit 和 maxfd 的调优 /**
func Test1() {
	arr := make([]*Req, 0)
	for i := 0; i < 1000; i++ {
		join.Add(1)
		req := &Req{}
		x := req.Url("http://localhost:8080/get/1").
			Method(http.MethodGet).
			Header("Connection", "Keep-Alive").
			Header("Content-Type", "application/json").
			Timeout(30).
			Build()
		arr = append(arr, x)
	}
	for _, req := range arr {
		go runAndPrint(req)
	}
	join.Wait()
}

// 并发请求
func runAndPrint(r *Req) {
	defer join.Done()
	r.Go()
	//fmt.Println(.Body())
}
