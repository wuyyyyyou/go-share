package httputils

import (
	"io"
	"net/http"
	"net/url"
)

// HttpRequest 封装了一个HTTP请求，并返回响应体的流
func HttpRequest(client *http.Client, urlStr, method string, queries map[string]string, requestBody io.Reader, headers, cookies map[string]string) (int, io.ReadCloser, error) {
	// 解析URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return 0, nil, err
	}

	// 添加查询参数
	params := parsedURL.Query()
	for key, value := range queries {
		params.Add(key, value)
	}
	parsedURL.RawQuery = params.Encode()

	// 创建请求
	req, err := http.NewRequest(method, parsedURL.String(), requestBody)
	if err != nil {
		return 0, nil, err
	}

	// 添加头部
	for key, value := range headers {
		req.Header.Add(key, value)
	}

	// 添加Cookie
	for name, value := range cookies {
		req.AddCookie(&http.Cookie{Name: name, Value: value})
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, err
	}

	// 返回状态码和响应体流
	return resp.StatusCode, resp.Body, nil
}
