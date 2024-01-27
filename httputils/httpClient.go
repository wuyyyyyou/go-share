package httputils

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/wuyyyyyou/go-share/ioutils"
)

type HttpClient struct {
	url     string
	queries map[string]string
	headers map[string]string
	cookies map[string]string

	client *http.Client

	requestBody io.Reader
	response    *http.Response
}

func NewHttpClient(url string) *HttpClient {
	return &HttpClient{
		url:     url,
		queries: make(map[string]string),
		headers: make(map[string]string),
		cookies: make(map[string]string),

		client: &http.Client{},
	}
}

/*
下面是添加请求参数、头部、Cookie的方法
*/

func (h *HttpClient) SetQueries(queries map[string]string) {
	h.queries = queries
}

func (h *HttpClient) SetQuery(key string, value string) {
	h.queries[key] = value
}

func (h *HttpClient) SetHeaders(headers map[string]string) {
	h.headers = headers
}

func (h *HttpClient) SetHeader(key string, value string) {
	h.headers[key] = value
}

func (h *HttpClient) SetCookies(cookies map[string]string) {
	h.cookies = cookies
}

func (h *HttpClient) SetCookie(key string, value string) {
	h.cookies[key] = value
}

/*
设置*http.Client相关参数，如超时时间等
*/

// SetTimeout 设置超时时间
func (h *HttpClient) SetTimeout(time time.Duration) {
	h.client.Timeout = time
}

/*
下面用户设置不同格式的请求体
*/

// SetJsonBody 设置请求体为JSON格式
// 输入json对应的结构体或者map[string]any
func (h *HttpClient) SetJsonBody(jsonData any) error {
	jsonBytes, err := json.Marshal(jsonData)
	if err != nil {
		return err
	}

	h.requestBody = bytes.NewReader(jsonBytes)
	return nil
}

// SetFormBody 设置请求体为表单格式
// 输入 map[string]string 和 map[string]*os.File
func (h *HttpClient) SetFormBody(formData map[string]string, formFile map[string]*os.File) error {
	// 创建一个buffer用于存储请求体
	var requestBody bytes.Buffer
	bodyWriter := multipart.NewWriter(&requestBody)
	defer ioutils.CloseQuietly(bodyWriter)

	// 添加文件数据
	for fileKey, file := range formFile {
		// 添加一个文件
		fileWriter, err := bodyWriter.CreateFormFile(fileKey, file.Name())
		if err != nil {
			return err
		}

		// 将文件内容写入到请求体中
		_, err = io.Copy(fileWriter, file)
		if err != nil {
			return err
		}
		ioutils.CloseQuietly(file)
	}

	// 添加表单数据
	for key, value := range formData {
		err := bodyWriter.WriteField(key, value)
		if err != nil {
			return err
		}
	}

	// 设置请求头Content-Type
	h.SetHeader("Content-Type", bodyWriter.FormDataContentType())
	// 设置请求体
	h.requestBody = &requestBody

	return nil
}

/*
下面的代码用于正式开始http请求
*/

func (h *HttpClient) httpRequest(method string) error {
	parse, err := url.Parse(h.url)
	if err != nil {
		return err
	}

	// 设置查询参数
	params := parse.Query()
	for key, value := range h.queries {
		params.Set(key, value)
	}
	parse.RawQuery = params.Encode()

	h.url = parse.String()

	// 创建请求
	req, err := http.NewRequest(method, h.url, h.requestBody)
	if err != nil {
		return err
	}

	// 设置头部
	for key, value := range h.headers {
		req.Header.Set(key, value)
	}

	// 设置Cookie
	for key, value := range h.cookies {
		req.AddCookie(&http.Cookie{Name: key, Value: value})
	}

	// 发送请求
	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}

	h.response = resp

	return nil
}

// Get 发送GET请求
func (h *HttpClient) Get() error {
	return h.httpRequest("GET")
}

// Post 发送POST请求
func (h *HttpClient) Post() error {
	return h.httpRequest("POST")
}

/*
完成请求后，进行的一些操作
*/

func (h *HttpClient) GetResponse() *http.Response {
	return h.response
}

func (h *HttpClient) GetResponseStatusCode() int {
	return h.response.StatusCode
}

func (h *HttpClient) GetResponseHeader() http.Header {
	return h.response.Header
}

func (h *HttpClient) GetBodyBytes() ([]byte, error) {
	return io.ReadAll(h.response.Body)
}

func (h *HttpClient) GetBodyString() (string, error) {
	bodyBytes, err := io.ReadAll(h.response.Body)
	if err != nil {
		return "", err
	}
	return string(bodyBytes), nil
}

/*
清理资源操作
*/

func (h *HttpClient) Close() {
	if h.response != nil {
		ioutils.CloseQuietly(h.response.Body)
	}
}
