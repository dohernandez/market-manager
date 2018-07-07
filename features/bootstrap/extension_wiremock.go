package bootstrap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/DATA-DOG/godog"
)

// WireMock is a struct to interact with WireMock container.
type WireMock struct {
	Response  *Response `json:"response"`
	Request   *Request  `json:"request"`
	serverURI string
}

// RegisterWireMockExtension Register the WireMock management extension
func RegisterWireMockExtension(s *godog.Suite, serverURI string) *WireMock {
	wireMockServer := NewWireMock(serverURI)

	s.BeforeScenario(func(interface{}) {
		wireMockServer.Reset()
	})

	return wireMockServer
}

// NewWireMock creates and returns a new WireMock.
func NewWireMock(serverURI string) *WireMock {
	return &WireMock{
		serverURI: serverURI,
	}
}

// SetWireMockResponse adds a response to the object.
func (w *WireMock) SetWireMockResponse(response *Response) {
	w.Response = response
}

// SetWireMockRequest adds a request to the object.
func (w *WireMock) SetWireMockRequest(request *Request) {
	w.Request = request
}

// Send submits the mock to the WireMockServer.
func (w *WireMock) Send() (err error) {
	mJSON, err := json.Marshal(w)
	if err != nil {
		return
	}

	contentReader := bytes.NewReader(mJSON)
	req, err := http.NewRequest("POST", w.serverURI+"/__admin/mappings", contentReader)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := execRequest(req)
	if err != nil {
		return
	}

	return assertStatusEquals(resp, http.StatusCreated)
}

// Reset erases all Mocks configured in the WireMockServer.
func (w *WireMock) Reset() (err error) {
	req, err := http.NewRequest("POST", w.serverURI+"/__admin/mappings/reset", nil)
	if err != nil {
		return
	}

	resp, err := execRequest(req)
	if err != nil {
		return
	}

	if err := assertStatusEquals(resp, http.StatusOK); err != nil {
		return err
	}

	return
}

func execRequest(req *http.Request) (resp *http.Response, err error) {
	client := &http.Client{}

	return client.Do(req)
}

// Response is the configuration to the Response of a mock.
type Response struct {
	Status int               `json:"status"`
	Body   string            `json:"body"`
	Header map[string]string `json:"headers"`
}

// NewWireMockResponse creates and returns a new response.
func NewWireMockResponse() *Response {
	resp := &Response{}
	resp.Header = map[string]string{}

	return resp
}

// SetStatus adds a status code to the response.
func (r *Response) SetStatus(status int) {
	r.Status = status
}

// SetBody adds a body to the response.
func (r *Response) SetBody(body string) {
	r.Body = body
}

// SetHeader adds a header to the response.
func (r *Response) SetHeader(key, value string) {
	r.Header[key] = value
}

// SetTextBody sets it as text body and add a header in the response.
func (r *Response) SetTextBody(body string) error {
	r.SetBody(body)
	r.SetHeader("Content-Type", "text/plain;charset=utf-8")

	return nil
}

// SetBodyJSON marshals an object into a JSON string, sets it as body and add a header in the response.
func (r *Response) SetJSONBody(body interface{}) error {
	jsonData, err := json.Marshal(body)
	if err != nil {
		return err
	}

	r.SetBody(string(jsonData))
	r.SetHeader("Content-Type", "application/json")

	return nil
}

// Request is the configuration to the request of a mock.
type Request struct {
	Method string `json:"method"`
	URL    string `json:"url"`
	Body   string `json:"body,omitempty"`
}

// NewWireMockRequest creates and returns a new request.
func NewWireMockRequest() *Request {
	return &Request{}
}

// SetMethod adds a method to the request.
func (r *Request) SetMethod(method string) {
	r.Method = method
}

// SetURL adds a url to the request.
func (r *Request) SetURL(url string) {
	r.URL = url
}

func assertStatusEquals(response *http.Response, statusCode int) error {
	if response.StatusCode == statusCode {
		return nil
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("invalid response. Error when trying to retrieve the body. StatusCode: %d", response.StatusCode)
	}

	return fmt.Errorf("invalid response. StatusCode: %d Body: %s", response.StatusCode, body)
}
