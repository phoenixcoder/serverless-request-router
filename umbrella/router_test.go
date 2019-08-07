package main

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

const (
	testBody        = "Test Body"
	testError       = "Test Error"
	mRRespBody      = "Http Client Response"
	testUrl         = "http://example.com/some/endpoint"
	testContentType = "Test Content Type"
)

type mockHttpClient struct {
	mock.Mock
}

type mockReader struct {
	mock.Mock
}

func (m *mockHttpClient) Post(url string, contentType string, body io.Reader) (*http.Response, error) {
	args := m.Called(url, contentType, body)
	return args[0].(*http.Response), args.Error(1)
}

func (m *mockReader) Read(text []byte) (int, error) {
	args := m.Called(text)
	return args.Int(0), args.Error(1)
}

func (m *mockReader) Close() error {
	return nil
}

func makeMockHCResp(body string) *http.Response {
	return &http.Response{
		Status:     http.StatusText(http.StatusOK),
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}

func TestRouteRequest(t *testing.T) {
	mResp := makeMockHCResp(mRRespBody)
	mResp.Header.Set(contentTypeHeader, testContentType)
	mHttpClient := new(mockHttpClient)
	mHttpClient.On("Post", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.Anything).Return(mResp, nil)
	testResp, err := routeRequest(testUrl, testBody, mHttpClient)

	assert.Nil(t, err)
	assert.Equal(t, testResp.StatusCode, mResp.StatusCode)
	assert.Equal(t, testResp.Body, mRRespBody)
	assert.Equal(t, testResp.Headers[contentTypeHeader], mResp.Header.Get(contentTypeHeader))
}

func TestRouteRequestFailedResponse(t *testing.T) {
	mResp := makeMockHCResp(mRRespBody)
	mHttpClient := new(mockHttpClient)
	mErr := errors.New(testError)
	mHttpClient.On("Post", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.Anything).Return(mResp, mErr)
	testResp, err := routeRequest(testUrl, testBody, mHttpClient)

	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), mErr.Error())
	assert.Equal(t, testResp.StatusCode, http.StatusInternalServerError)
	assert.Equal(t, testResp.Body, internalErrRespMsg+" ("+http.StatusText(http.StatusInternalServerError)+")")
}

func TestRouteRequestFailedReader(t *testing.T) {
	mResp := makeMockHCResp(mRRespBody)
	mHttpClient := new(mockHttpClient)
	mReader := new(mockReader)
	mResp.Body = mReader
	mErr := errors.New(testError)
	mHttpClient.On("Post", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.Anything).Return(mResp, nil)
	mReader.On("Read", mock.Anything).Return(0, mErr)
	testResp, err := routeRequest(testUrl, testBody, mHttpClient)

	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), mErr.Error())
	assert.Equal(t, testResp.StatusCode, http.StatusInternalServerError)
	assert.Equal(t, testResp.Body, internalErrRespMsg+" ("+http.StatusText(http.StatusInternalServerError)+")")
}
