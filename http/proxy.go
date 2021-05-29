package http

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
)

// ProxyUrl 优先使用传入的代理
func (req *Req) ProxyUrl(proxyUrl string) *Req {
	if proxyUrl == "" && len(req.proxyArray) > 0 {
		random := rand.Intn(len(req.proxyArray))
		proxyUrl =  req.proxyArray[random]
	} else {
		fmt.Println("can not found proxy ...")
		return req
	}
	uri, err := url.Parse("http://" + proxyUrl)
	if err != nil {
		fmt.Println("proxylist.txt url parse err ", err)
		return req
	}
	if req.transport == nil{
		transport := &http.Transport{Proxy: http.ProxyURL(uri)}
		req.transport = transport
	}else{
		req.transport.Proxy = http.ProxyURL(uri)
	}
	return req
}

// Proxy 重载带参数的方法
func (req *Req) Proxy() *Req {
      req.ProxyUrl("")
      return req
}
// ImportProxy 必须提前配置好proxy代理文件
func (req *Req)ImportProxy() *Req{
	path , err := os.Getwd()
	if err != nil {
		fmt.Println("pwd is err",err)
		return req
	}
	file, err := ioutil.ReadFile( path + "/proxy1.txt")
	if err != nil {
		fmt.Println("read proxylist.txt error err = ",err)
		return req
	}
	scanner := bufio.NewScanner(bytes.NewReader(file))
	for scanner.Scan() {
		if scanner.Text() != "" {
			req.proxyArray = append(req.proxyArray, scanner.Text())
		}
	}
	fmt.Println("import proxy success count " , len(req.proxyArray))
	return req
}


