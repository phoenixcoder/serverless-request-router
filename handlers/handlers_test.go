package handlers

import (
	"errors"
	"github.com/phoenixcoder/serverless-request-router/router"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

const (
	testProxyUrl         = "https://testurl.com/"
	testProxyBody        = "Test Proxy Body"
	testProxyContentType = "application/json"
)

type mockHttpClient struct {
	mock.Mock
}

type mockReader struct {
	mock.Mock
}

func (m *mockReader) Read(p []byte) (int, error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func (m *mockReader) Close() error {
	args := m.Called()

	return args.Error(0)
}

func (m *mockHttpClient) Post(url string, contentType string, body io.Reader) (*http.Response, error) {
	args := m.Called(url, contentType, body)
	return args[0].(*http.Response), args.Error(1)
}

func TestProxyHandler(t *testing.T) {
	testCtx := make(router.ContextMap)
	testTask := make(router.TaskMap)
	testResp := &http.Response{
		Body: ioutil.NopCloser(strings.NewReader(testProxyBody)),
	}
	mHttpClient := new(mockHttpClient)

	testCtx[requestUrl] = testProxyUrl
	testCtx[contentTypeHeader] = testProxyContentType

	testHandler := NewProxyHandler(mHttpClient)

	mHttpClient.On("Post", mock.Anything, mock.Anything, mock.Anything).Return(testResp, nil)

	testHandler.Execute(&testCtx, &testTask)

	assert.Equal(t, testTask[TaskBody], testProxyBody)
	assert.Nil(t, testTask[ErrorKey])
	assert.Equal(t, testCtx[requestUrl], testProxyUrl)
	assert.Equal(t, testCtx[contentTypeHeader], testProxyContentType)

	mHttpClient.AssertNumberOfCalls(t, "Post", 1)
}

func TestProxyHandlerPostError(t *testing.T) {
	testError := errors.New("Post Error")
	testCtx := make(router.ContextMap)
	testTask := make(router.TaskMap)
	testResp := &http.Response{}
	mHttpClient := new(mockHttpClient)
	mReader := new(mockReader)
	testCtx[requestUrl] = testProxyUrl
	testCtx[contentTypeHeader] = testProxyContentType
	testResp.Body = mReader
	testHandler := NewProxyHandler(mHttpClient)

	mHttpClient.On("Post", mock.Anything, mock.Anything, mock.Anything).Return(testResp, testError)
	mReader.On("Read", mock.Anything).Return(0, nil)

	testHandler.Execute(&testCtx, &testTask)

	assert.False(t, testHandler.Before(nil, nil))
	assert.Nil(t, testTask[TaskBody])
	assert.NotNil(t, testTask[ErrorKey])
	assert.Equal(t, testTask[ErrorKey], testError)
	assert.Equal(t, testCtx[requestUrl], testProxyUrl)
	assert.Equal(t, testCtx[contentTypeHeader], testProxyContentType)

	mHttpClient.AssertNumberOfCalls(t, "Post", 1)
	mReader.AssertNotCalled(t, "Read", mock.Anything)
}

func TestProxyHandlerReadingError(t *testing.T) {
	testError := errors.New("Reading Error")
	testCtx := make(router.ContextMap)
	testTask := make(router.TaskMap)
	testResp := &http.Response{}
	mHttpClient := new(mockHttpClient)
	mReader := new(mockReader)
	testCtx[requestUrl] = testProxyUrl
	testCtx[contentTypeHeader] = testProxyContentType
	testResp.Body = mReader
	testHandler := NewProxyHandler(mHttpClient)

	mHttpClient.On("Post", mock.Anything, mock.Anything, mock.Anything).Return(testResp, nil)
	mReader.On("Read", mock.Anything).Return(0, testError)

	testHandler.Execute(&testCtx, &testTask)

	assert.False(t, testHandler.Before(nil, nil))
	assert.Nil(t, testTask[TaskBody])
	assert.NotNil(t, testTask[ErrorKey])
	assert.Equal(t, testTask[ErrorKey], testError)
	assert.Equal(t, testCtx[requestUrl], testProxyUrl)
	assert.Equal(t, testCtx[contentTypeHeader], testProxyContentType)

	mHttpClient.AssertNumberOfCalls(t, "Post", 1)
	mReader.AssertNumberOfCalls(t, "Read", 1)
}

func TestProxyHandlerNoRequestUrl(t *testing.T) {
	testCtx := make(router.ContextMap)
	testTask := make(router.TaskMap)
	testResp := &http.Response{}
	mHttpClient := new(mockHttpClient)
	mReader := new(mockReader)
	testCtx[contentTypeHeader] = testProxyContentType
	testResp.Body = mReader
	testHandler := NewProxyHandler(mHttpClient)

	mHttpClient.On("Post", mock.Anything, mock.Anything, mock.Anything).Return(testResp, nil)
	mReader.On("Read", mock.Anything).Return(0, nil)

	testHandler.Execute(&testCtx, &testTask)

	assert.False(t, testHandler.Before(nil, nil))
	assert.Nil(t, testTask[TaskBody])
	assert.Nil(t, testCtx[requestUrl])
	assert.NotNil(t, testTask[ErrorKey])
	assert.Equal(t, testCtx[contentTypeHeader], testProxyContentType)
	assert.Equal(t, testTask[ErrorKey].(error).Error(), testHandler.errMsg)

	mHttpClient.AssertNotCalled(t, "Post", mock.Anything, mock.Anything, mock.Anything)
	mReader.AssertNotCalled(t, "Read", mock.Anything)
}
