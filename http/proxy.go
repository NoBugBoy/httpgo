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
func (r *Req) ProxyUrl(proxyUrl string) *Req {
	if r.proxy == nil && r.client == nil {
		r.client = &http.Client{}
	}
	if proxyUrl == "" && len(r.proxyArray) > 0 {
		random := rand.Intn(len(r.proxyArray))
		proxyUrl =  r.proxyArray[random]
	} else {
		fmt.Println("can not found proxy ...")
		return r
	}
	uri, err := url.Parse("http://" + proxyUrl)
	if err != nil {
		fmt.Println("proxylist.txt url parse err ", err)
		return r
	}
	transport := &http.Transport{Proxy: http.ProxyURL(uri)}
	r.client.Transport = transport
	return r
}

// Proxy 重载带参数的方法
func (r *Req) Proxy() *Req {
      r.ProxyUrl("")
      return r
}
// ImportProxy 必须提前配置好proxy代理文件
func (r *Req)ImportProxy() *Req{
	path , err := os.Getwd()
	if err != nil {
		fmt.Println("pwd is err",err)
		return r
	}
	file, err := ioutil.ReadFile( path + "/proxy1.txt")
	if err != nil {
		fmt.Println("read proxylist.txt error err = ",err)
		return r
	}
	scanner := bufio.NewScanner(bytes.NewReader(file))
	for scanner.Scan() {
		if scanner.Text() != "" {
			r.proxyArray = append(r.proxyArray, scanner.Text())
		}
	}
	fmt.Println("import proxy success count " , len(r.proxyArray))
	return r
}


