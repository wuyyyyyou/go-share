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
	"github.com/wuyyyyyou/go-share/share"
)

type HttpRequest struct {
	url     string
	queries map[string]string
	headers map[string]string
	cookies map[string]string

	client *http.Client

	requestBody io.Reader
	response    *http.Response
}

func NewHttpRequest(url string) *HttpRequest {
	return &HttpRequest{
		url:     share.EnsureHttpPrefix(url),
		queries: make(map[string]string),
		headers: make(map[string]string),
		cookies: make(map[string]string),

		client: &http.Client{},
	}
}

/*
下面是添加请求参数、头部、Cookie的方法
*/

func (h *HttpRequest) SetQueries(queries map[string]string) {
	h.queries = queries
}

func (h *HttpRequest) SetQuery(key string, value string) {
	h.queries[key] = value
}

func (h *HttpRequest) SetHeaders(headers map[string]string) {
	h.headers = headers
}

func (h *HttpRequest) SetHeader(key string, value string) {
	h.headers[key] = value
}

func (h *HttpRequest) SetCookies(cookies map[string]string) {
	h.cookies = cookies
}

func (h *HttpRequest) SetCookie(key string, value string) {
	h.cookies[key] = value
}

/*
设置*http.Client相关参数，如超时时间等
*/

// SetTimeout 设置超时时间
func (h *HttpRequest) SetTimeout(time time.Duration) {
	h.client.Timeout = time
}

/*
下面用户设置不同格式的请求体
*/

// SetJsonBody 设置请求体为JSON格式
// 输入json对应的结构体或者map[string]any
func (h *HttpRequest) SetJsonBody(jsonData any) error {
	jsonBytes, err := json.Marshal(jsonData)
	if err != nil {
		return err
	}

	h.requestBody = bytes.NewReader(jsonBytes)
	return nil
}

// SetFormBody 设置请求体为表单格式
// formData包含了form表单的键值对
// fromFile包含了要输入的文件，键值对中的key对应的是form表单中的name，value对应的是文件的路径
func (h *HttpRequest) SetFormBody(formData map[string]string, formFile map[string]string) error {
	// 创建一个buffer用于存储请求体
	var requestBody bytes.Buffer
	bodyWriter := multipart.NewWriter(&requestBody)
	defer ioutils.CloseQuietly(bodyWriter)

	// 添加文件数据
	for fileKey, filePath := range formFile {
		err := func() error {
			file, err := os.Open(filePath)
			defer ioutils.CloseQuietly(file)
			if err != nil {
				return err
			}

			fileWriter, err := bodyWriter.CreateFormFile(fileKey, file.Name())
			if err != nil {
				return err
			}

			// 将文件内容写入到请求体中
			_, err = io.Copy(fileWriter, file)
			if err != nil {
				return err
			}

			return nil
		}()

		if err != nil {
			return err
		}
	}

	// 添加表单数据
	for key, value := range formData {
		err := bodyWriter.WriteField(key, value)
		if err != nil {
			return err
		}
	}

	// 根据 formFile 的长度设置 Content-Type
	if len(formFile) > 0 {
		// 设置请求头Content-Type为multipart/form-data
		h.SetHeader("Content-Type", bodyWriter.FormDataContentType())
	} else {
		// 设置请求头Content-Type为application/x-www-form-urlencoded
		h.SetHeader("Content-Type", "application/x-www-form-urlencoded")
	}

	// 设置请求体
	h.requestBody = &requestBody

	return nil
}

/*
下面的代码用于正式开始http请求
*/

func (h *HttpRequest) httpRequest(method string) error {
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
func (h *HttpRequest) Get() error {
	return h.httpRequest("GET")
}

// Post 发送POST请求
func (h *HttpRequest) Post() error {
	return h.httpRequest("POST")
}

/*
完成请求后，进行的一些操作
*/

func (h *HttpRequest) GetResponse() *http.Response {
	return h.response
}

func (h *HttpRequest) GetResponseStatusCode() int {
	return h.response.StatusCode
}

func (h *HttpRequest) GetResponseHeader() http.Header {
	return h.response.Header
}

func (h *HttpRequest) GetBodyBytes() ([]byte, error) {
	return io.ReadAll(h.response.Body)
}

func (h *HttpRequest) GetBodyString() (string, error) {
	bodyBytes, err := io.ReadAll(h.response.Body)
	if err != nil {
		return "", err
	}
	return string(bodyBytes), nil
}

/*
清理资源操作
*/

func (h *HttpRequest) Close() {
	if h.response != nil {
		ioutils.CloseQuietly(h.response.Body)
	}
}
