package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"
)

type ClientInterface interface {
	SendRequest(request RequestInterface) (ResponseInterface, error)
}

type RequestInterface interface {
	GetMethod() string
	GetUri() string
	GetHeaders() map[string]string
	GetBody() io.Reader
}

type ResponseInterface interface {
	GetStatusCode() int
	GetHeaders() map[string]string
	GetBody() []byte
}

type Client struct {
	client   *http.Client
	baseURL  string
	headers  map[string]string
	timeout  time.Duration
}

func NewClient() *Client {
	return &Client{
		client:  &http.Client{Timeout: 30 * time.Second},
		headers: make(map[string]string),
		timeout: 30 * time.Second,
	}
}

func (c *Client) SetBaseURL(baseURL string) *Client {
	c.baseURL = baseURL
	return c
}

func (c *Client) SetTimeout(timeout time.Duration) *Client {
	c.timeout = timeout
	c.client.Timeout = timeout
	return c
}

func (c *Client) SetHeader(key, value string) *Client {
	c.headers[key] = value
	return c
}

func (c *Client) Get(ctx context.Context, url string) (*Response, error) {
	return c.Do(ctx, "GET", url, nil, nil)
}

func (c *Client) Post(ctx context.Context, url string, body interface{}) (*Response, error) {
	return c.Do(ctx, "POST", url, body, nil)
}

func (c *Client) Put(ctx context.Context, url string, body interface{}) (*Response, error) {
	return c.Do(ctx, "PUT", url, body, nil)
}

func (c *Client) Delete(ctx context.Context, url string) (*Response, error) {
	return c.Do(ctx, "DELETE", url, nil, nil)
}

func (c *Client) Patch(ctx context.Context, url string, body interface{}) (*Response, error) {
	return c.Do(ctx, "PATCH", url, body, nil)
}

func (c *Client) Do(ctx context.Context, method string, url string, body interface{}, headers map[string]string) (*Response, error) {
	fullURL := c.baseURL + url

	var bodyReader io.Reader
	if body != nil {
		switch v := body.(type) {
		case string:
			bodyReader = bytes.NewBufferString(v)
		case []byte:
			bodyReader = bytes.NewBuffer(v)
		default:
			jsonBytes, err := json.Marshal(body)
			if err != nil {
				return nil, err
			}
			bodyReader = bytes.NewBuffer(jsonBytes)
			c.headers["Content-Type"] = "application/json"
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return nil, err
	}

	for key, value := range c.headers {
		req.Header.Set(key, value)
	}
	if headers != nil {
		for key, value := range headers {
			req.Header.Set(key, value)
		}
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	responseHeaders := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			responseHeaders[key] = values[0]
		}
	}

	return &Response{
		statusCode: resp.StatusCode,
		headers:    responseHeaders,
		body:      respBody,
	}, nil
}

type Response struct {
	statusCode int
	headers    map[string]string
	body       []byte
}

func (r *Response) StatusCode() int {
	return r.statusCode
}

func (r *Response) Headers() map[string]string {
	return r.headers
}

func (r *Response) Body() []byte {
	return r.body
}

func (r *Response) BodyString() string {
	return string(r.body)
}

func (r *Response) Header(name string) string {
	return r.headers[name]
}

func (r *Response) Ok() bool {
	return r.statusCode >= 200 && r.statusCode < 300
}

func (r *Response) IsRedirect() bool {
	return r.statusCode >= 300 && r.statusCode < 400
}

func (r *Response) IsClientError() bool {
	return r.statusCode >= 400 && r.statusCode < 500
}

func (r *Response) IsServerError() bool {
	return r.statusCode >= 500
}

func (r *Response) Json(v interface{}) error {
	return json.Unmarshal(r.body, v)
}

func (r *Response) Query() url.Values {
	q, _ := url.ParseQuery(string(r.body))
	return q
}

type GzipClient struct {
	client *Client
}

func NewGzipClient() *GzipClient {
	return &GzipClient{client: NewClient()}
}

func (c *GzipClient) SetBaseURL(url string) *GzipClient {
	c.client.SetBaseURL(url)
	return c
}

func (c *GzipClient) SetTimeout(t time.Duration) *GzipClient {
	c.client.SetTimeout(t)
	return c
}

func (c *GzipClient) Do(ctx context.Context, method, url string, body interface{}) (*Response, error) {
	req, _ := http.NewRequestWithContext(ctx, method, c.client.baseURL+url, nil)
	req.Header.Set("Accept-Encoding", "gzip")
	return c.client.Do(ctx, method, url, body, nil)
}

type MockClient struct {
	responses []*Response
	requests  []*Request
	index    int
}

func NewMockClient() *MockClient {
	return &MockClient{
		responses: make([]*Response, 0),
		requests:  make([]*Request, 0),
		index:    0,
	}
}

func (m *MockClient) AddResponse(resp *Response) {
	m.responses = append(m.responses, resp)
}

func (m *MockClient) Do(ctx context.Context, method, url string, body interface{}) (*Response, error) {
	m.requests = append(m.requests, &Request{Method: method, URL: url, Body: body})

	if m.index >= len(m.responses) {
		return &Response{statusCode: 500, body: []byte("No more mock responses")}, nil
	}

	resp := m.responses[m.index]
	m.index++
	return resp, nil
}

func (m *MockClient) GetRequests() []*Request {
	return m.requests
}

type Request struct {
	Method string
	URL    string
	Body   interface{}
}
