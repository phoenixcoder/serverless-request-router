package router

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

const (
	testBody        = "Test Body"
	testError       = "Test Error"
	mRRespBody      = "Http Client Response"
	testUrl         = "http://example.com/some/endpoint"
	testContentType = "Test Content Type"
	testCtxContent  = "Test Ctx Content"
	testCtxKey      = "Test Ctx Key"
	testTaskContent = "Test Task Content"
	testTaskKey     = "Test Task Key"

	testBeforeKey = "Test Before Key"
	testExecKey   = "Test Exec Key"
	testAfterKey  = "Test After Key"

	testBeforeCtxContent = "Test Before Ctx Content"
	testExecCtxContent   = "Test Exec Ctx Content"
	testAfterCtxContent  = "Test After Ctx Content"

	testBeforeTaskContent = "Test Before Task Content"
	testExecTaskContent   = "Test Exec Task Content"
	testAfterTaskContent  = "Test After Task Content"
)

type mockHandler struct {
	mock.Mock
}

func (m *mockHandler) Before(context *ContextMap, task *TaskMap) bool {
	args := m.Called(context, task)

	return args.Bool(0)
}

func (m *mockHandler) Execute(context *ContextMap, task *TaskMap) {
	m.Called(context, task)
}

func (m *mockHandler) After(context *ContextMap, task *TaskMap) {
	m.Called(context, task)
}

type mockRequestHelper struct {
	count int
	task  TaskMap
}

func (m *mockRequestHelper) mockRequestAdapter(request interface{}) TaskMap {
	m.count++
	return m.task
}

type mockResponseHelper struct {
	count    int
	response interface{}
}

func (m *mockResponseHelper) mockResponseAdapter(task *TaskMap) interface{} {
	m.count++
	return m.response
}

type mockContextHelper struct {
	count   int
	context ContextMap
}

func (m *mockContextHelper) contextCreator() ContextMap {
	m.count++
	return m.context
}

type mapModifyingHandler struct {
	mock.Mock
	ctx  ContextMap
	task TaskMap
}

func (c *mapModifyingHandler) Before(context *ContextMap, task *TaskMap) bool {
	args := c.Called(context, task)
	(*context)[testBeforeKey] = c.ctx[testBeforeKey]
	(*task)[testBeforeKey] = c.task[testBeforeKey]
	return args.Bool(0)
}

func (c *mapModifyingHandler) Execute(context *ContextMap, task *TaskMap) {
	c.Called(context, task)
	(*context)[testExecKey] = c.ctx[testExecKey]
	(*task)[testExecKey] = c.task[testExecKey]
}

func (c *mapModifyingHandler) After(context *ContextMap, task *TaskMap) {
	c.Called(context, task)
	(*context)[testAfterKey] = c.ctx[testAfterKey]
	(*task)[testAfterKey] = c.task[testAfterKey]
}

type mockRequest interface{}
type mockResponse interface{}

func TestRouter(t *testing.T) {
	mockHandler1 := new(mockHandler)
	mockHandler2 := new(mockHandler)
	mockReq := new(mockRequest)
	mockCtxHelper := &mockContextHelper{
		context: make(ContextMap),
	}
	mockReqHelper := &mockRequestHelper{
		task: *new(TaskMap),
	}
	mockRespHelper := &mockResponseHelper{
		response: testBody,
	}
	mockHandler1.On("Before", mock.Anything, mock.Anything).Return(false)
	mockHandler1.On("Execute", mock.Anything, mock.Anything)
	mockHandler1.On("After", mock.Anything, mock.Anything)
	mockHandler2.On("Before", mock.Anything, mock.Anything).Return(false)
	mockHandler2.On("Execute", mock.Anything, mock.Anything)
	mockHandler2.On("After", mock.Anything, mock.Anything)

	testRouter := newRouter(mockCtxHelper.contextCreator, mockReqHelper.mockRequestAdapter, mockRespHelper.mockResponseAdapter, mockHandler1, mockHandler2)
	testRes := testRouter.Handle(mockReq)

	assert.NotNil(t, testRes)
	assert.Equal(t, testRes.(string), testBody)
	assert.Equal(t, mockCtxHelper.count, 1)
	assert.Equal(t, mockReqHelper.count, 1)
	assert.Equal(t, mockRespHelper.count, 1)
	mockHandler1.AssertNumberOfCalls(t, "Before", 1)
	mockHandler1.AssertNumberOfCalls(t, "After", 1)
	mockHandler1.AssertNotCalled(t, "Execute", mock.Anything, mock.Anything)
	mockHandler2.AssertNumberOfCalls(t, "Before", 1)
	mockHandler2.AssertNumberOfCalls(t, "After", 1)
	mockHandler2.AssertNumberOfCalls(t, "Execute", 1)
}

func TestRouterModifyMaps(t *testing.T) {
	mockReq := new(mockRequest)
	testCtx := ContextMap{
		testCtxKey: testCtxContent,
	}
	testTask := TaskMap{
		testTaskKey: testTaskContent,
	}
	mapHandler := &mapModifyingHandler{
		ctx: ContextMap{
			testBeforeKey: testBeforeCtxContent,
			testExecKey:   testExecCtxContent,
			testAfterKey:  testAfterCtxContent,
		},
		task: TaskMap{
			testBeforeKey: testBeforeTaskContent,
			testExecKey:   testExecTaskContent,
			testAfterKey:  testAfterTaskContent,
		},
	}
	mockCtxHelper := &mockContextHelper{
		context: testCtx,
	}
	mockReqHelper := &mockRequestHelper{
		task: testTask,
	}
	mockRespHelper := &mockResponseHelper{
		response: testBody,
	}

	mapHandler.On("Before", mock.Anything, mock.Anything).Return(false)
	mapHandler.On("Execute", mock.Anything, mock.Anything)
	mapHandler.On("After", mock.Anything, mock.Anything)

	testRouter := newRouter(mockCtxHelper.contextCreator, mockReqHelper.mockRequestAdapter, mockRespHelper.mockResponseAdapter, mapHandler)
	testRes := testRouter.Handle(mockReq)

	assert.NotNil(t, testRes)
	assert.Equal(t, testRes.(string), testBody)

	mapHandler.AssertNumberOfCalls(t, "Before", 1)
	mapHandler.AssertNumberOfCalls(t, "Execute", 1)
	mapHandler.AssertNumberOfCalls(t, "After", 1)

	assert.Equal(t, testCtx[testCtxKey], testCtxContent)
	assert.Equal(t, mapHandler.ctx[testBeforeKey], testCtx[testBeforeKey])
	assert.Equal(t, mapHandler.ctx[testExecKey], testCtx[testExecKey])
	assert.Equal(t, mapHandler.ctx[testAfterKey], testCtx[testAfterKey])

	assert.Equal(t, testTask[testTaskKey], testTaskContent)
	assert.Equal(t, mapHandler.task[testBeforeKey], testTask[testBeforeKey])
	assert.Equal(t, mapHandler.task[testExecKey], testTask[testExecKey])
	assert.Equal(t, mapHandler.task[testAfterKey], testTask[testAfterKey])
}

func TestRouterTurnAround(t *testing.T) {
	mockHandler1 := new(mockHandler)
	mockHandler2 := new(mockHandler)
	mockHandler3 := new(mockHandler)
	mockReq := new(mockRequest)
	mockCtxHelper := &mockContextHelper{
		context: make(ContextMap),
	}
	mockReqHelper := &mockRequestHelper{
		task: *new(TaskMap),
	}
	mockRespHelper := &mockResponseHelper{
		response: testBody,
	}
	mockHandler1.On("Before", mock.Anything, mock.Anything).Return(false)
	mockHandler1.On("Execute", mock.Anything, mock.Anything)
	mockHandler1.On("After", mock.Anything, mock.Anything)

	mockHandler2.On("Before", mock.Anything, mock.Anything).Return(true)
	mockHandler2.On("Execute", mock.Anything, mock.Anything)
	mockHandler2.On("After", mock.Anything, mock.Anything)

	mockHandler3.On("Before", mock.Anything, mock.Anything)
	mockHandler3.On("Execute", mock.Anything, mock.Anything)
	mockHandler3.On("After", mock.Anything, mock.Anything)

	testRouter := newRouter(mockCtxHelper.contextCreator, mockReqHelper.mockRequestAdapter, mockRespHelper.mockResponseAdapter, mockHandler1, mockHandler2, mockHandler3)
	testRes := testRouter.Handle(mockReq)

	assert.NotNil(t, testRes)
	assert.Equal(t, testRes.(string), testBody)
	assert.Equal(t, mockCtxHelper.count, 1)
	assert.Equal(t, mockReqHelper.count, 1)
	assert.Equal(t, mockRespHelper.count, 1)

	mockHandler1.AssertNumberOfCalls(t, "Before", 1)
	mockHandler1.AssertNotCalled(t, "Execute", 1)
	mockHandler1.AssertNumberOfCalls(t, "After", 1)

	mockHandler2.AssertNumberOfCalls(t, "Before", 1)
	mockHandler2.AssertNumberOfCalls(t, "Execute", 1)
	mockHandler2.AssertNumberOfCalls(t, "After", 1)

	mockHandler3.AssertNotCalled(t, "Before", mock.Anything, mock.Anything)
	mockHandler3.AssertNotCalled(t, "After", mock.Anything, mock.Anything)
	mockHandler3.AssertNotCalled(t, "Execute", mock.Anything, mock.Anything)

}
