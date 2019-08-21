package handlers

import (
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

const (
	requestUrl        = "request-url"
	contentTypeHeader = "content-type"

	// TODO Move to a separate package.
	forbiddenErrRespMsg   = "uh uh uh...you didn't say the magic word."
	internalErrRespMsg    = "Sorry...we uh...messed up."
	funcNotFoundErrMsgFmt = "We're embarrassed for you, but we don't know a '%s'. Try these instead:\n'%s'"
	logMsg                = "%s, %s"
	ErrorKey              = "error"
	TaskBody              = "body"
)

// TODO Move to a separate package.
func setInternalErrCode(task *TaskMap, reason string) {
	setErredStatusCode(task, internalErrRespMsg, reason, http.StatusInternalServerError)
}

// TODO Move to a separate package.
func setForbiddenErrCode(task *TaskMap) {
	setErredStatusCode(task, forbiddenErrRespMsg, "You're just not allowed.", http.StatusForbidden)
}

// TODO Move to a separate package.
func setErredStatusCode(task *TaskMap, msg string, reason string, statusCode int) {
	(*task)["StatusCode"] = statusCode
	(*task)["Body"] = msg + " (" + http.StatusText(statusCode) + ")"
	log.Printf(logMsg, (*task)["Body"], reason)
}

type httpClientInterface interface {
	Post(url string, contentType string, body io.Reader) (*http.Response, error)
}

// ProxyHandler is a wrapper for the http client and process that
// forwards on the request to the intended service.
type ProxyHandler struct {
	errMsg            string
	requestUrl        string
	contentTypeHeader string
	client            httpClientInterface
}

// NewProxyHandler is a factory method for creating the proxy handler with
// a separately configured http client.
func NewProxyHandler(client httpClientInterface) ProxyHandler {
	return ProxyHandler{
		errMsg: "Request url was not provided when proxying.",
		client: client,
	}
}

// Before method that does nothing.
func (p *ProxyHandler) Before(context *ContextMap, task *TaskMap) bool {
	return false
}

// Execute method that inspects the context for a request url and sends
// an http request to that url. It sends the response of the request
// back in the task.
func (p *ProxyHandler) Execute(context *ContextMap, task *TaskMap) {
	body := (*task)[TaskBody]
	bodyStr, _ := body.(string)
	contentType := (*context)[contentTypeHeader]
	contentTypeStr, _ := contentType.(string)
	requestUrl, requestUrlOk := (*context)[requestUrl]
	requestUrlStr, requestUrlStrOk := requestUrl.(string)
	requestUrlOk = requestUrlOk && requestUrlStrOk
	if requestUrlOk {
		routedResp, err := p.client.Post(requestUrlStr, contentTypeStr, strings.NewReader(bodyStr))
		if err != nil {
			(*task)[ErrorKey] = err
			return
		}

		routedRespBody, err := ioutil.ReadAll(routedResp.Body)
		if err != nil {
			(*task)[ErrorKey] = err
			return
		}

		(*task)[TaskBody] = string(routedRespBody)
		return
	}

	(*task)[ErrorKey] = errors.New(p.errMsg)
}

// After method that does nothing.
func (p *ProxyHandler) After(context *ContextMap, task *TaskMap) {}
