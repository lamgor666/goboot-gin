package HttpClient

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/lamgor666/goboot-common/util/castx"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"
)

type client struct {
	requestUrl           string
	headers              map[string]string
	caCertPool           *x509.CertPool
	certificates         []tls.Certificate
	skipServerHostVerify bool
	timeout              time.Duration
}

type multipartEntry interface {
	IsMultipart() bool
}

type normalPart struct {
	data map[string]string
}

func (p normalPart) IsMultipart() bool {
	return true
}

type filePart struct {
	formFieldName  string
	clientFileName string
	buf            []byte
}

func (p filePart) IsMultipart() bool {
	return true
}

var httpErrors = map[int]string {
	400: "Bad Request",
	401: "Unauthorized",
	402: "Payment Required",
	403: "Forbidden",
	404: "Not Found",
	405: "Method Not Allowed",
	406: "Not Acceptable",
	407: "Proxy Authentication",
	408: "Request Timeout",
	409: "Conflict",
	410: "Gone",
	411: "Length Required",
	412: "Precondition Failed",
	413: "Request Entity Too Large",
	414: "Request-URI Too Long",
	415: "Unsupported Media Type",
	416: "Requested Range Not Satisfiable",
	417: "Expectation Failed",
	418: "I'm a teapot",
	421: "Misdirected Request",
	422: "Unprocessable Entity",
	423: "Locked",
	424: "Failed Dependency",
	425: "Too Early",
	426: "Upgrade Required",
	449: "Retry With",
	451: "Unavailable For Legal Reasons",
	500: "Internal Server Error",
	501: "Not Implemented",
	502: "Bad Gateway",
	503: "Service Unavailable",
	504: "Gateway Timeout",
	505: "HTTP Version Not Supported",
	506: "Variant Also Negotiates",
	507: "Insufficient Storage",
	509: "Bandwidth Limit Exceeded",
	510: "Not Extended",
	600: "Unparseable Response Headers",
}

func NewNormalPart(data map[string]string) normalPart {
	if len(data) < 1 {
		return normalPart{}
	}

	return normalPart{data: data}
}

func NewFilePart(formFieldName, clientFileName, fpath string) filePart {
	if formFieldName == "" || clientFileName == "" || fpath == "" {
		return filePart{}
	}

	buf, err := ioutil.ReadFile(fpath)

	if err != nil || len(buf) < 1 {
		return filePart{}
	}

	return filePart{
		formFieldName:  formFieldName,
		clientFileName: clientFileName,
		buf:            buf,
	}
}

func NewFilePartFromBuffer(formFieldName, clientFileName string, buf []byte) filePart {
	if formFieldName == "" || clientFileName == "" || len(buf) < 1 {
		return filePart{}
	}

	return filePart{
		formFieldName:  formFieldName,
		clientFileName: clientFileName,
		buf:            buf,
	}
}

func New(requestUrl string) client {
	return client{
		requestUrl:           requestUrl,
		headers:              map[string]string{},
		skipServerHostVerify: true,
		timeout:              15 * time.Second,
	}
}

func (c client) AddHeader(headerName string, headerValue string) client {
	headerName = c.headerNameToUcwords(headerName)
	c.headers[headerName] = headerValue
	return c
}

func (c client) SetHeaders(headers map[string]string) client {
	if len(headers) < 1 {
		return c
	}

	for headerName, headerValue := range headers {
		c.AddHeader(headerName, headerValue)
	}

	return c
}

func (c client) EnableSslVerify(certpem, keypem string, skipServerHostVerify ...bool) client {
	if certpem == "" || keypem == "" {
		return c
	}

	var certBuf, keyBuf []byte

	if strings.HasPrefix(certpem, "file://") {
		fpath := strings.TrimPrefix(certpem, "file://")
		certBuf, _ = ioutil.ReadFile(fpath)
	} else {
		certBuf = []byte(certpem)
	}

	if strings.HasPrefix(keypem, "file://") {
		fpath := strings.TrimPrefix(certpem, "file://")
		keyBuf, _ = ioutil.ReadFile(fpath)
	} else {
		keyBuf = []byte(certpem)
	}

	if len(certBuf) < 1 || len(keyBuf) < 1 {
		return c
	}

	cert, err := tls.X509KeyPair(certBuf, keyBuf)

	if err != nil {
		return c
	}

	c.certificates = append(c.certificates, cert)
	_skipServerHostVerify := true

	if len(skipServerHostVerify) > 0 {
		_skipServerHostVerify = skipServerHostVerify[0]
	}

	c.skipServerHostVerify = _skipServerHostVerify
	return c
}

func (c client) EnableSslVerifyWithCacert(cacert, certpem, keypem string) client {
	c.EnableSslVerify(certpem, keypem)

	if cacert == "" || len(c.certificates) < 1 {
		return c
	}

	var buf []byte

	if strings.HasPrefix(cacert, "file://") {
		fpath := strings.TrimPrefix(cacert, "file://")
		buf, _ = ioutil.ReadFile(fpath)
	} else {
		buf = []byte(cacert)
	}

	if len(buf) < 1 {
		return c
	}

	pool := x509.NewCertPool()

	if pool.AppendCertsFromPEM(buf) {
		c.caCertPool = pool
		c.skipServerHostVerify = false
	}

	return c
}

func (c client) WithTimeout(arg0 interface{}) client {
	var timeout time.Duration

	if d1, ok := arg0.(time.Duration); ok {
		timeout = d1
	} else if n1, ok := arg0.(int); ok && n1 > 0 {
		timeout = time.Duration(castx.ToInt64(n1)) * time.Second
	} else if s1, ok := arg0.(string); ok && s1 != "" {
		if d1, err := time.ParseDuration(s1); err == nil {
			timeout = d1
		}
	}

	if timeout >= time.Second {
		c.timeout = timeout
	}

	return c
}

func (c client) Get(args ...map[string]string) ([]byte, error) {
	queryString := map[string]string{}

	for _, arg := range args {
		for key, value := range arg {
			if _, ok := queryString[key]; ok {
				continue
			}

			queryString[key] = value
		}
	}

	if len(queryString) > 0 {
		urlValues := make(url.Values, len(queryString))

		for key, value := range queryString {
			urlValues.Set(key, value)
		}

		if strings.Contains(c.requestUrl, "?") {
			c.requestUrl += "&" + urlValues.Encode()
		} else {
			c.requestUrl += "?" + urlValues.Encode()
		}
	}

	return c.sendRequest("GET", "", nil)
}

func (c client) Post(args ...map[string]string) ([]byte, error) {
	formData := map[string]string{}

	for _, arg := range args {
		for key, value := range arg {
			if _, ok := formData[key]; ok {
				continue
			}

			formData[key] = value
		}
	}

	if len(formData) < 1 {
		return c.sendRequest("POST", "application/x-www-form-urlencoded", nil)
	}

	urlValues := make(url.Values, len(formData))

	for key, value := range formData {
		urlValues.Set(key, value)
	}

	contentType := "application/x-www-form-urlencoded"
	body := strings.NewReader(urlValues.Encode())
	return c.sendRequest("POST", contentType, body)
}

func (c client) PostMultipartForm(parts ...multipartEntry) ([]byte, error) {
	if len(parts) < 1 {
		return c.Post()
	}

	var normalParts []normalPart
	var fileParts []filePart

	for _, entry := range parts {
		if p, ok := entry.(normalPart); ok {
			normalParts = append(normalParts, p)
			continue
		}

		if p, ok := entry.(filePart); ok {
			fileParts = append(fileParts, p)
		}
	}

	buf := &bytes.Buffer{}
	body := multipart.NewWriter(buf)
	defer body.Close()

	for _, np := range normalParts {
		if len(np.data) < 1 {
			continue
		}

		for key, value := range np.data {
			if err := body.WriteField(key, value); err != nil {
				return nil, err
			}
		}
	}

	for _, fp := range fileParts {
		var writer io.Writer
		writer, err := body.CreateFormFile(fp.formFieldName, fp.clientFileName)

		if err != nil {
			return nil, err
		}

		if _, err := io.Copy(writer, bytes.NewReader(fp.buf)); err != nil {
			return nil, err
		}
	}

	return c.sendRequest("POST", body.FormDataContentType(), buf)
}

func (c client) PostXml(xml string) ([]byte, error) {
	return c.sendRequest("POST", "text/xml", strings.NewReader(xml))
}

func (c client) PostJson(json string) ([]byte, error) {
	return c.sendRequest("POST", "application/json", strings.NewReader(json))
}

func (c client) PutXml(xml string) ([]byte, error) {
	return c.sendRequest("PUT", "text/xml", strings.NewReader(xml))
}

func (c client) PutJson(json string) ([]byte, error) {
	return c.sendRequest("PUT", "application/json", strings.NewReader(json))
}

func (c client) PatchXml(xml string) ([]byte, error) {
	return c.sendRequest("PATCH", "text/xml", strings.NewReader(xml))
}

func (c client) PatchJson(json string) ([]byte, error) {
	return c.sendRequest("PATCH", "application/json", strings.NewReader(json))
}

func (c client) Delete(args ...map[string]string) ([]byte, error) {
	queryString := map[string]string{}

	for _, arg := range args {
		for key, value := range arg {
			if _, ok := queryString[key]; ok {
				continue
			}

			queryString[key] = value
		}
	}

	if len(queryString) > 0 {
		urlValues := make(url.Values, len(queryString))

		for key, value := range queryString {
			urlValues.Set(key, value)
		}

		if strings.Contains(c.requestUrl, "?") {
			c.requestUrl += "&" + urlValues.Encode()
		} else {
			c.requestUrl += "?" + urlValues.Encode()
		}
	}

	return c.sendRequest("DELETE", "", nil)
}

func (c client) sendRequest(method, contentType string, body io.Reader) (buf []byte, err error) {
	if contentType != "" {
		c.AddHeader("Content-Type", contentType)
	}

	req, err := c.buildRequest(method, body)

	if err != nil {
		return
	}

	resp, err := c.buildClient().Do(req)

	if err != nil {
		return
	}

	defer resp.Body.Close()
	buf, err = ioutil.ReadAll(resp.Body)
	statusCode := resp.StatusCode

	if statusCode >= 400 {
		reason, ok := httpErrors[statusCode]
		var errorTips string

		if ok {
			errorTips = fmt.Sprintf("http error %d %s", statusCode, reason)
		} else {
			errorTips = fmt.Sprintf("http error %d", statusCode)
		}

		if len(buf) > 0 {
			errorTips += "\nresponse: " + string(buf)
		}

		buf = nil
		err = errors.New(errorTips)
	}

	return
}

func (c client) buildClient() (client *http.Client) {
	transport := &http.Transport{}
	schema := strings.ToLower(c.substringBefore(c.requestUrl, "://"))

	if schema != "https" {
		client = &http.Client{Transport: transport}

		if c.timeout >= time.Second {
			client.Timeout = c.timeout
		}

		return
	}

	tlsConfig := &tls.Config{}

	if len(c.certificates) > 0 {
		tlsConfig.Certificates = c.certificates
	}

	if c.skipServerHostVerify || c.caCertPool == nil {
		tlsConfig.InsecureSkipVerify = true
	} else {
		tlsConfig.RootCAs = c.caCertPool
	}

	transport.TLSClientConfig = tlsConfig
	client = &http.Client{Transport: transport}

	if c.timeout >= time.Second {
		client.Timeout = c.timeout
	}

	return
}

func (c client) buildRequest(method string, body io.Reader) (req *http.Request, err error) {
	req, err = http.NewRequest(method, c.requestUrl, body)

	if err != nil {
		return
	}

	for headerName, headerValue := range c.headers {
		req.Header.Set(headerName, headerValue)
	}

	return
}

func (c client) toStringMapString(arg0 interface{}) map[string]string {
	if arg0 == nil {
		return map[string]string{}
	}

	if map1, ok := arg0.(map[string]string); ok {
		return map1
	}

	rt := reflect.TypeOf(arg0)
	rv := reflect.ValueOf(arg0)

	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
		rv = rv.Elem()
	}

	if rt.Kind() != reflect.Map {
		return map[string]string{}
	}

	map1 := map[string]string{}
	iter := rv.MapRange()

	for iter.Next() {
		key, ok := iter.Key().Interface().(string)

		if !ok || key == "" {
			continue
		}

		value, ok := iter.Value().Interface().(string)

		if !ok {
			continue
		}

		map1[key] = value
	}

	return map1
}

func (c client) substringBefore(str, delimiter string) string {
	idx := strings.Index(str, delimiter)

	if idx < 1 {
		return str
	}

	return str[:idx]
}

func (c client) headerNameToUcwords(headerName string) string {
	parts := strings.Split(headerName, "-")

	for i, p := range parts {
		if len(p) < 2 {
			parts[i] = strings.ToUpper(p)
			continue
		}

		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}

	return strings.Join(parts, "-")
}
