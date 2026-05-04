package request

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

type Request struct {
	http.Request
}

func NewRequest(r *http.Request) *Request {
	return &Request{*r}
}

func (r *Request) Context() context.Context {
	return r.Request.Context()
}

func (r *Request) IsAJAX() bool {
	return r.Header.Get("X-Requested-With") == "XMLHttpRequest"
}

func (r *Request) IsJson() bool {
	return strings.Contains(r.Header.Get("Accept"), "application/json")
}

func (r *Request) GetClientIP() string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ",")[0]
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	return r.RemoteAddr
}

func (r *Request) GetPreferredLanguage(defaultLang string) string {
	acceptLang := r.Header.Get("Accept-Language")
	if acceptLang == "" {
		return defaultLang
	}

	languages := strings.Split(acceptLang, ",")
	for _, lang := range languages {
		lang = strings.Split(lang, ";")[0]
		lang = strings.TrimSpace(lang)
		if len(lang) == 2 {
			return lang
		}
	}
	return defaultLang
}

func (r *Request) GetHost() string {
	if host := r.Header.Get("X-Forwarded-Host"); host != "" {
		return host
	}
	return r.Host
}

func (r *Request) GetScheme() string {
	if r.Header.Get("X-Forwarded-Proto") != "" {
		return r.Header.Get("X-Forwarded-Proto")
	}
	if r.TLS != nil {
		return "https"
	}
	return "http"
}

func (r *Request) GetUri() string {
	if r.RequestURI != "" {
		return r.RequestURI
	}
	return r.URL.RequestURI()
}

func (r *Request) GetFullUrl() string {
	return r.GetScheme() + "://" + r.GetHost() + r.GetUri()
}

func (r *Request) GetBearerToken() string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ""
	}

	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}
	return parts[1]
}

func (r *Request) GetBasicAuth() (username, password string, ok bool) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return "", "", false
	}

	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "basic" {
		return "", "", false
	}

	payload, err := base64Decode(parts[1])
	if err != nil {
		return "", "", false
	}

	username, password, ok = split2(payload, ":")
	return
}

func base64Decode(s string) ([]byte, error) {
	return decodeBase64(s)
}

func decodeBase64(s string) ([]byte, error) {
	const encodeStd = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	decoder := make([]byte, 256)
	for i := range decoder {
		decoder[i] = 0xFF
	}
	for i := 0; i < len(encodeStd); i++ {
		decoder[encodeStd[i]] = byte(i)
	}

	remain := len(s) % 4
	if remain > 0 {
		s += strings.Repeat("=", 4-remain)
	}

	result := make([]byte, len(s)*6/8)
	resultIdx := 0
	chunk := 0
	bits := 0

	for _, c := range s {
		if c == '=' {
			break
		}
		val := decoder[byte(c)]
		if val == 0xFF {
			continue
		}
		chunk = chunk<<6 | int(val)
		bits += 6
		if bits >= 8 {
			bits -= 8
			result[resultIdx] = byte(chunk >> bits)
			resultIdx++
			chunk &= (1<<bits) - 1
		}
	}
	return result[:resultIdx], nil
}

func split2(s []byte, sep byte) (before, after string, ok bool) {
	i := indexByte(s, sep)
	if i < 0 {
		return "", "", false
	}
	return string(s[:i]), string(s[i+1:]), true
}

func indexByte(s []byte, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

func (r *Request) GetAcceptEncoding() []string {
	enc := r.Header.Get("Accept-Encoding")
	if enc == "" {
		return nil
	}
	return strings.Split(enc, ",")
}

func (r *Request) GetContentType() string {
	return r.Header.Get("Content-Type")
}

func (r *Request) IsContentType(ct string) bool {
	return strings.Contains(r.GetContentType(), ct)
}

func (r *Request) GetReferer() string {
	return r.Header.Get("Referer")
}

func (r *Request) GetUserAgent() string {
	return r.Header.Get("User-Agent")
}

func (r *Request) HasHeader(name string) bool {
	return r.Header.Get(name) != ""
}

func (r *Request) QueryString() string {
	if r.URL.RawQuery != "" {
		return "?" + r.URL.RawQuery
	}
	return ""
}

func (r *Request) QueryParams() url.Values {
	return r.URL.Query()
}

func (r *Request) GetQueryArray(key string) []string {
	if values, ok := r.URL.Query()[key]; ok {
		return values
	}
	return []string{}
}

func (r *Request) HasQuery(key string) bool {
	_, ok := r.URL.Query()[key]
	return ok
}

func (r *Request) GetIntQuery(key string, defaultValue int64) int64 {
	if val := r.URL.Query().Get(key); val != "" {
		if i, ok := parseInt64(val); ok {
			return i
		}
	}
	return defaultValue
}

func (r *Request) GetFloatQuery(key string, defaultValue float64) float64 {
	if val := r.URL.Query().Get(key); val != "" {
		if f, ok := parseFloat(val); ok {
			return f
		}
	}
	return defaultValue
}

func (r *Request) GetBoolQuery(key string, defaultValue bool) bool {
	if val := r.URL.Query().Get(key); val != "" {
		val = strings.ToLower(val)
		if val == "true" || val == "1" || val == "yes" || val == "on" {
			return true
		}
		if val == "false" || val == "0" || val == "no" || val == "off" {
			return false
		}
	}
	return defaultValue
}

var numberRegex = regexp.MustCompile(`^-?\d+(\.\d+)?`)

func parseInt64(s string) (int64, bool) {
	m := numberRegex.FindString(s)
	if m == "" {
		return 0, false
	}
	n := int64(0)
	negative := false
	for _, c := range m {
		if c == '-' {
			negative = true
			continue
		}
		n = n*10 + int64(c-'0')
	}
	if negative {
		n = -n
	}
	return n, true
}

func parseFloat(s string) (float64, bool) {
	m := numberRegex.FindString(s)
	if m == "" {
		return 0, false
	}
	f := float64(0)
	decimal := false
	divisor := 1.0
	negative := false
	for _, c := range m {
		if c == '-' {
			negative = true
			continue
		}
		if c == '.' {
			decimal = true
			continue
		}
		if !decimal {
			f = f*10 + float64(c-'0')
		} else {
			divisor *= 10
			f += float64(c-'0') / divisor
		}
	}
	if negative {
		f = -f
	}
	return f, true
}

func (r *Request) ParseForm() error {
	if r.Form != nil {
		return nil
	}
	return r.ParseMultipartForm(32 << 20)
}

func (r *Request) ParseQueryForm() error {
	if r.Form != nil {
		return nil
	}
	r.Form = make(url.Values)
	query := r.URL.Query().Encode()
	if query != "" {
		r.Form, _ = url.ParseQuery(query)
	}
	return nil
}

func (r *Request) GetFormValue(key string) string {
	if r.Form == nil {
		r.ParseForm()
	}
	return r.FormValue(key)
}

func (r *Request) GetFormArray(key string) []string {
	if r.Form == nil {
		r.ParseForm()
	}
	if values, ok := r.Form[key]; ok {
		return values
	}
	return []string{}
}

func (r *Request) HasForm(key string) bool {
	if r.Form == nil {
		r.ParseForm()
	}
	_, ok := r.Form[key]
	return ok
}

func (r *Request) GetFormFile(key string) ([]byte, error) {
	if r.MultipartForm == nil {
		return nil, nil
	}
	file, _, err := r.FormFile(key)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return io.ReadAll(file)
}

type XmlMap map[string]interface{}

func (r *Request) ParseXml(v interface{}) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return xml.Unmarshal(body, v)
}

func (r *Request) ParseJson(v interface{}) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

func (r *Request) GetAllHeaders() map[string][]string {
	return r.Header
}

func (r *Request) HeaderContains(name, value string) bool {
	return strings.Contains(r.Header.Get(name), value)
}

func (r *Request) MethodIs(methods ...string) bool {
	for _, m := range methods {
		if r.Method == m {
			return true
		}
	}
	return false
}

func (r *Request) IsGet() bool    { return r.Method == http.MethodGet }
func (r *Request) IsPost() bool   { return r.Method == http.MethodPost }
func (r *Request) IsPut() bool    { return r.Method == http.MethodPut }
func (r *Request) IsDelete() bool { return r.Method == http.MethodDelete }
func (r *Request) IsPatch() bool  { return r.Method == http.MethodPatch }
func (r *Request) IsHead() bool   { return r.Method == http.MethodHead }
func (r *Request) IsOptions() bool { return r.Method == http.MethodOptions }

func (r *Request) WantsJson() bool {
	accept := r.Header.Get("Accept")
	if accept == "" {
		return false
	}
	return strings.Contains(accept, "application/json")
}

func (r *Request) WantsXml() bool {
	accept := r.Header.Get("Accept")
	if accept == "" {
		return false
	}
	return strings.Contains(accept, "application/xml") || strings.Contains(accept, "text/xml")
}

func (r *Request) WantsHtml() bool {
	accept := r.Header.Get("Accept")
	if accept == "" {
		return false
	}
	return strings.Contains(accept, "text/html")
}
