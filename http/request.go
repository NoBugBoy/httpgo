package http

import (
	"bytes"
	"errors"
	"fmt"
	json "github.com/json-iterator/go"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var methods = map[string]string{
	"GET": http.MethodGet,
	"POST": http.MethodPost,
	"OPTIONS": http.MethodOptions,
	"DELETE": http.MethodDelete,
	"PUT": http.MethodPut,
}

type Query map[string]interface{}

// Req core object
type Req struct {
	client *http.Client // http client
	method string //请求方式
	url *url.URL // 请求路径
	proxy *url.URL // 本次使用的代理
	proxyArray []string // 代理列表
	pathQuery bool //是否是路径请求参数
	timeout time.Duration //超时时间（s）
	retry int //重试次数
	header http.Header //http header
	Request *http.Request // 请求对象
	Response *http.Response // 响应对象
	result string //最后的返回值，为了正常关闭io
	params map[string]interface{} //请求参数一律 key : value
}
// Header 设置请求头，多个调用多次即可
func (r *Req) Header(key string,value string) *Req {
	if r.header == nil {
		header := http.Header{}
		header.Set(key,value)
		r.header = header
	}else{
		r.header.Set(key,value)
	}
	return r
}

// Url 设置请求地址 格式为 http://开头
func  (r *Req) Url(httpPath string) *Req {
	_url, err := url.Parse(httpPath)
	if err != nil{
		fmt.Println("http url parse error , check url format ",err)
		return r
	}
	if str := strings.Contains(_url.String(), "?"); str{
		//路径携带？意味着是pathQuery
		r.pathQuery = true
	}
	r.url = _url
	return r
}

// Method 设置请求方式建议使用 http.MethodGet 方式
func (r *Req) Method(method string) *Req {

	m := methods[strings.ToUpper(method)]
	if m == "" {
		fmt.Println("can not find method ",method)
		return r
	}
	r.method = m
	return r
}

func (r *Req) Params(m Query) *Req {
	r.params = m
	return r
}

// ParseParams deep parse get params
func ParseParams(v interface{}) string{
	if k, ok := v.(string); ok {
		return k
	}else if k, ok := v.(int); ok{
		return strconv.Itoa(k)
	}else if k, ok := v.(float32); ok{
		return fmt.Sprintf("%.2f", k)
	}else if k, ok := v.(float64); ok{
		return fmt.Sprintf("%.2f", k)
	}else if k, ok := v.([]interface{}); ok{
		for i := range k {
			ParseParams(i)
		}
	}else{
		re := reflect.TypeOf(v)
		for i := 0; i < re.NumField(); i++ {
			f := re.Field(i)
			value := reflect.ValueOf(f)
			ParseParams(value)
		}
	}
	return ""
}
func  buildGetParam(params map[string]interface{}) string {
	buff := bytes.Buffer{}
	buff.WriteString("?")
	for k, v := range params {
		buff.WriteString(k)
		buff.WriteString("=")
		buff.WriteString(ParseParams(v))
		buff.WriteString("&")
	}
	return string(buff.Next(buff.Len()-1))
}
func buildJson(params map[string]interface{}) []byte{
	var b, _ = json.Marshal(params)
	return b
}

// Timeout 设置请求超时时间
func (r *Req) Timeout(second int) *Req{
	if second < 0 {
		fmt.Println("timeout must gt 0 ")
		second = 3
	}
	if second > 60 {
		second = 60
	}
	r.timeout = time.Duration(second) * time.Second
	return r
}
// Retry 设置重试次数最大3次
func (r *Req) Retry(count int) *Req{
	if count < 0 {
		count = 0
	}
	if count > 3 {
		count = 3
	}
	r.retry = count
	return r
}

// Build 预构建（方便统一调用Go）
func (r *Req) Build() *Req{
	var realpath string
	var data []byte = nil
	if r.method == http.MethodGet && r.pathQuery {
		realpath = r.url.String()
	}else if r.method == http.MethodGet && r.params != nil{
		realpath = r.url.String() + buildGetParam(r.params)
	}else{
		realpath = r.url.String()
		data =  buildJson(r.params)
   }
	req , err := http.NewRequest(r.method,realpath,bytes.NewReader(data))
	if err != nil {
		fmt.Println("http new err = ",err)
		return r
	}
	// add header
	req.Header = r.header
	// add request time out
	if r.timeout.Seconds() <= 0{
		//默认是3s
		r.timeout = time.Duration(3) * time.Second
	}
	// check new httpclient
	if r.client == nil{
		r.client = &http.Client{}
	}

	r.client.Timeout = r.timeout
	r.Request = req
	return r
}

// Go 实际调用发送请求
func (r *Req) Go() *Req {
	// 未预构建，自动构建请求
	if r.Request == nil{
		r.Build()
	}
	//关闭连接
	defer func(*http.Request) {
		if r.Request != nil && r.Request.Body != nil {
			err := r.Request.Body.Close()
			if err != nil {
				fmt.Println("close request err",err)
			}
		}
	}(r.Request)
	// 关闭之后无法再次读取,[]byte len = 0
	defer func(*http.Response) {
		if r.Response != nil && r.Response.Body != nil{
			err := r.Response.Body.Close()
			if err != nil {
				fmt.Println("关闭response连接 err",err)
			}
		}
	}(r.Response)
	count := 1
	// do while ?
	for{
		res,err := r.client.Do(r.Request)
		if err != nil && r.retry == 0{
			fmt.Println("http call err = ",err)
			return r
		}else if err != nil && r.retry > 0 {
			fmt.Println("http call err retrying ...",count , "\n err = " , err)
			r.retry --
			count ++
			if r.retry == 0 {
				return r
			}
		}else{
			// ok
			r.Response = res
			b, _ := ioutil.ReadAll(r.Response.Body)
			r.result = string(b)
			break
		}
	}
	return r
}

// Body 直接获取返回值
func (r *Req) Body() (string,error) {
	if r.Response == nil || r.result == ""{
		return "",errors.New("empty response")
	}
	return r.result,nil
}


