package pkg

type Response struct {
	body       string
	statusCode int
	headers    map[string]string
}

func NewResponse(body string, statusCode int) *Response {
	return &Response{
		body:       body,
		statusCode: statusCode,
		headers:    make(map[string]string),
	}
}

func (r *Response) Body() string {
	return r.body
}

func (r *Response) StatusCode() int {
	return r.statusCode
}

func (r *Response) Headers() map[string]string {
	return r.headers
}

func (r *Response) WithHeader(key, value string) *Response {
	r.headers[key] = value
	return r
}

func (r *Response) JSON(data interface{}) *Response {
	r.headers["Content-Type"] = "application/json"
	r.body = toJSON(data)
	return r
}