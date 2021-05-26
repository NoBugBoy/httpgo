package http

import (
	"bytes"
	"fmt"
	json "github.com/json-iterator/go"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
)

var methods = map[string]string{
	"GET": http.MethodGet,
	"POST": http.MethodPost,
	"OPTIONS": http.MethodOptions,
	"DELETE": http.MethodDelete,
	"PUT": http.MethodPut,
}
type Query map[string]interface{}

type Req struct {
	client *http.Client
	method string
	url *url.URL
	param string
	header http.Header
	request *http.Request
	response *http.Response
	params map[string]interface{}
}

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

func  (r *Req) Url(httpPath string) *Req {
	_url, err := url.Parse(httpPath)
	if err != nil{
		fmt.Println("http url parse error ",err)
		os.Exit(1)
	}
	r.url = _url
	return r
}

func (r *Req) Method(method string) *Req {
	m := methods[strings.ToUpper(method)]
	if m == "" {
		fmt.Println("can not find method ",method)
		os.Exit(1)
	}
	r.method = m
	return r
}

func (r *Req) Params( m Query) *Req {
	r.params = m
	return r
}
func ParseParams(v interface{}) string{
	if k, ok := v.(string); ok {
		return k
	}else if k, ok := v.(int); ok{
		return strconv.Itoa(k)
	}else if k, ok := v.(float32); ok{
		return fmt.Sprintf("%.2f", k)
	}else if k, ok := v.(float64); ok{
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
func (r *Req) Go() *Req {
	r.client = &http.Client{}
	var realpath string
	var json []byte = nil
	if r.method == http.MethodGet && r.params==nil {
		realpath = r.url.String()
	}else if r.method == http.MethodGet && r.params!=nil{
		realpath = r.url.String() + buildGetParam(r.params)
	}else{
		realpath = r.url.String()
		json =  buildJson(r.params)
	}
	req , err := http.NewRequest(r.method,realpath,bytes.NewReader(json))
	if err != nil {
		fmt.Println("http new err",err)
		os.Exit(1)
	}

	req.Header = r.header
	res,err := r.client.Do(req)
	if err != nil {
		fmt.Println("http call err",err)
		os.Exit(1)
	}
	r.response = res
	return r
}
func (r *Req) Body() string {
	body := r.response.Body
	b, _ := ioutil.ReadAll(body)
	return string(b)
}

