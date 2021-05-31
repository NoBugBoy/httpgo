package http

import (
	"bytes"
	"errors"
	"fmt"
	json "github.com/json-iterator/go"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)
//单例http.Client,性能和多例对比差别不大，单内存占有会降低少许
var httpClient *http.Client

func init()  {
	httpClient = &http.Client{}
}

var methods = map[string]string{
	"GET":     http.MethodGet,
	"POST":    http.MethodPost,
	"OPTIONS": http.MethodOptions,
	"DELETE":  http.MethodDelete,
	"PUT":     http.MethodPut,
}

type Query map[string]interface{}

// Req core object
type Req struct {
	client     *http.Client           // http client
	transport  *http.Transport        //调优使用
	method     string                 //请求方式
	url        *url.URL               // 请求路径
	proxy      *url.URL               // 本次使用的代理
	proxyArray []string               // 代理列表
	pathQuery  bool                   //是否是路径请求参数
	timeout    time.Duration          //超时时间（s）
	retry      int                    //重试次数
	chunk      bool                   // chunk模式，自己控制读取Response.body,这里不会自动关闭
	header     http.Header            //http header
	Request    *http.Request          // 请求对象
	Response   *http.Response         // 响应对象
	result     string                 //最后的返回值，为了正常关闭io
	params     map[string]interface{} //请求参数一律 key : value
	isReady    bool //是否已经构建
}

// Header 设置请求头，多个调用多次即可
func (req *Req) Header(key string, value string) *Req {
	if req.header == nil {
		header := http.Header{}
		header.Set(key, value)
		req.header = header
	} else {
		req.header.Set(key, value)
	}
	return req
}

// Chunk 开启chunk模式，自己控制读取Response.body,注意要手动关闭io
func (req *Req) Chunk() *Req {
	req.chunk = true
	return req
}

// TransportSetting 对连接调优
func (req *Req) TransportSetting() *http.Transport {
	if req.transport == nil{
		req.transport = &http.Transport{}
	}
	return req.transport
}

// Url 设置请求地址 格式为 http://开头
func (req *Req) Url(httpPath string) *Req {
	_url, err := url.Parse(httpPath)
	if err != nil {
		fmt.Println("http url parse error , check url format ", err)
		return req
	}
	if str := strings.Contains(_url.String(), "?"); str {
		//路径携带？意味着是pathQuery
		req.pathQuery = true
	}
	req.url = _url
	return req
}

// Method 设置请求方式建议使用 http.MethodGet 方式
func (req *Req) Method(method string) *Req {
	m := methods[strings.ToUpper(method)]
	if m == "" {
		fmt.Println("can not find method ", method)
		return req
	}
	req.method = m
	return req
}

// Params 设置请求参数 get post 等均可
func (req *Req) Params(m Query) *Req {
	req.params = m
	return req
}

// ParseParams deep parse get params
func parseParams(v interface{}) string {
	if k, ok := v.(string); ok {
		return k
	} else if k, ok := v.(int); ok {
		return strconv.Itoa(k)
	} else if k, ok := v.(float32); ok {
		return fmt.Sprintf("%.4f", k)
	} else if k, ok := v.(float64); ok {
		return fmt.Sprintf("%.4f", k)
	} else if k, ok := v.([]interface{}); ok {
		for i := range k {
			parseParams(i)
		}
	} else {
		re := reflect.TypeOf(v)
		for i := 0; i < re.NumField(); i++ {
			f := re.Field(i)
			value := reflect.ValueOf(f)
			parseParams(value)
		}
	}
	return ""
}
func BuildGetParam(params map[string]interface{}) string {
	buff := bytes.Buffer{}
	buff.WriteString("?")
	for k, v := range params {
		buff.WriteString(k)
		buff.WriteString("=")
		buff.WriteString(parseParams(v))
		buff.WriteString("&")
	}
	return string(buff.Next(buff.Len() - 1))
}
func buildJson(params map[string]interface{}) []byte {
	var b, _ = json.Marshal(params)
	return b
}

// Timeout 设置请求超时时间
func (req *Req) Timeout(second int) *Req {
	if second < 0 {
		fmt.Println("timeout must gt 0 ")
		second = 3
	}
	if second > 60 {
		second = 60
	}
	req.timeout = time.Duration(second) * time.Second
	return req
}

// Retry 设置重试次数最大3次
func (req *Req) Retry(count int) *Req {
	if count < 0 {
		count = 0
	}
	if count > 3 {
		count = 3
	}
	req.retry = count
	return req
}

// Build 预构建（方便统一调用Go）
func (req *Req) Build() *Req {
	var realpath string
	var data []byte = nil
	if req.method == http.MethodGet && req.pathQuery {
		realpath = req.url.String()
	} else if req.method == http.MethodGet && req.params != nil {
		realpath = req.url.String() + BuildGetParam(req.params)
	} else {
		realpath = req.url.String()
		data = buildJson(req.params)
	}
	httpRequest, err := http.NewRequest(req.method, realpath, bytes.NewReader(data))
	if err != nil {
		panic(err)
		return req
	}
	// add header
	httpRequest.Header = req.header
	// add request time out
	if req.timeout.Seconds() <= 0 {
		//默认是5s
		req.timeout = time.Duration(5) * time.Second
	}
	req.client = httpClient
	if req.transport != nil {
		req.client.Transport = req.transport
	}
	req.client.Timeout = req.timeout
	req.Request = httpRequest
	req.isReady = true
	return req
}

// Go 实际调用发送请求
func (req *Req) Go() *Req {
	// 未预构建，自动构建请求
	if !req.isReady {
		req.Build()
	}
	//关闭连接
	defer func(*http.Request) {
		if req.Request != nil && req.Request.Body != nil && !req.chunk {
			err := req.Request.Body.Close()
			if err != nil {
				fmt.Println("close request err", err)
			}
		}
	}(req.Request)
	// 关闭之后无法再次读取,[]byte len = 0
	defer func(*http.Response) {
		if req.Response != nil && req.Response.Body != nil && !req.chunk {
			_, _ = io.Copy(ioutil.Discard, req.Response.Body)
			err := req.Response.Body.Close()
			if err != nil {
				fmt.Println("关闭response连接 err", err)
			}
		}
	}(req.Response)
	count := 1
	// do while ?
	for {
		res, err := req.client.Do(req.Request)
		if err != nil && req.retry == 0 {
			fmt.Println("http call err = ", err)
			return req
		} else if err != nil && req.retry > 0 {
			fmt.Println("http call err retrying ...", count, "\n err = ", err)
			req.retry--
			count++
			if req.retry == 0 {
				return req
			}
		} else {
			// ok
			req.Response = res
			if !req.chunk {
				b, _ := ioutil.ReadAll(req.Response.Body)
				req.result = string(b)
			}
			break
		}
	}
	reInit(req)
	return req
}
func reInit(req *Req)  {
	req.isReady = false
	req.url = nil
	req.params = nil
	req.pathQuery = false
	req.chunk = false
	req.header = nil
}

// Body 直接获取返回值
func (req *Req) Body() (string, error) {
	if req.Response == nil || req.result == "" {
		return "", errors.New("empty response")
	}
	return req.result, nil
}
