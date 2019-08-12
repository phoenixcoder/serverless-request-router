package router

// Router manages a sequence of actions that occur to a request on its
// way into the service, and to the response on its way out. The sequence
// of actions are user-defined.
type Router struct {
	handler       Handler
	createContext ContextCreator
	adaptRequest  RequestAdapter
	adaptResponse ResponseAdapter
}

// ContextCreator is a factory method that generates a context map
// prior to the request getting handled.
type ContextCreator func() ContextMap

// RequestAdapter is a transformer method that converts an object of
// any type to a request map.
type RequestAdapter func(interface{}) TaskMap

// ResponseAdapter is a transformer method that converts a ResponseMap
// to a response object of any type.
type ResponseAdapter func(*TaskMap) interface{}

// ContextMap holds any environment information required during the course
// of a request.
type ContextMap map[string]interface{}

// TaskMap holds any information related to the request or response. This is
// dependent on what gets mapped from the RequestAdaptor passed in.
type TaskMap map[string]interface{}

// DefaultContextCreator is a packaged method that produces an empty
// map for the request path.
func DefaultContextCreator() ContextMap {
	return ContextMap(map[string]interface{}{})
}

// Handler is the interface for the  workhorse of the router. It defines
// what actions need to be taken at before, during, and after the request
// travels down a node of the path. It is a composition of
// the BeforeHandler, ExecuteHandler, and AfterHandler.
type Handler interface {
	BeforeHandler
	ExecuteHandler
	AfterHandler
}

// BeforeHandler is the interface that wraps the Before method.
type BeforeHandler interface {
	//   Before handles the request on the way into the service with
	//   the given context and/or task. The boolean return should indicate
	//   whether the service should stop processing the request or not.
	Before(context *ContextMap, task *TaskMap) bool
}

// AfterHandler is the interfaces that wraps the After method.
type AfterHandler interface {
	//   After handles the request on the way out of the service prior
	//   to the creation of a response.
	After(context *ContextMap, task *TaskMap)
}

// ExecuteHandler is the interface that wraps the Execute method.
type ExecuteHandler interface {
	// Execute handles the business logic to be performed because of the
	// request or any error handling that may need to occur.
	Execute(context *ContextMap, task *TaskMap)
}

// NewChainHandler is a creation method that takes in a list of handlers.
// Each Handler is wrapped by a ChainHandler, and linked to the next and previous
// Handler in the list, if they exist. The head of the doubly-linked list is returned
func NewChainHandler(handlers ...Handler) *ChainHandler {
	currHandler := &ChainHandler{}
	var prevHandler *ChainHandler
	head := currHandler
	n := len(handlers)
	for i := 0; i < n; i++ {
		currHandler.curr = handlers[i]

		if prevHandler != nil {
			currHandler.prev = prevHandler
			prevHandler.next = currHandler
		}

		if i < n-1 {
			prevHandler = currHandler
			currHandler = &ChainHandler{}
		}
	}

	return head
}

// ChainHandler is a node in a doubly-linked list of handlers responsible
// for executing the Before, Execute, and After methods of the current
// handler it wraps. It then is responsible for linking to the next
// link in the chain or the previous one.
type ChainHandler struct {
	next BeforeHandler
	prev AfterHandler
	curr Handler
}

// Before on the ChainHandler manages whether the next link in the request
// chain should be called or not. It takes in a context and task, and makes
// a call to a wrapped handler. If the wrapped handler indicates the request
// processing must stop or the end of the request chain has been reached,
// the current handler's Execute method is called. If the handler indicates
// a stop-condition has been reached, any stop or error information should be
// registered with the task and/or context.

// A boolean value is passed back from this method, but within the router
// returned by NewRouter, it is not used.
func (c *ChainHandler) Before(context *ContextMap, task *TaskMap) bool {
	stop := c.curr.Before(context, task)
	// If the next handler is empty, then we've reached the end of the
	// chain, and it's time to execute the intended logic. Otherwise,
	// execute the intended logic, which is meant to handle the error
	// case.
	if !stop && c.next != nil {
		c.next.Before(context, task)
		return false
	}

	c.Execute(context, task)
	return true
}

// Execute on the ChainHandler manages whether the current handler's
// Execute method is run. It then runs the current handler's After
// method regardless of the execution's results.
func (c *ChainHandler) Execute(context *ContextMap, task *TaskMap) {
	c.curr.Execute(context, task)
	c.After(context, task)
}

// After on the ChainHandler executes the current handler's After method
// is run. It then runs the previous handler's After method, if the
// request has not already reached the beginning of the list.
func (c *ChainHandler) After(context *ContextMap, task *TaskMap) {
	// Make sure we're not at the beginning of the chain.
	c.curr.After(context, task)
	if c.prev != nil {
		c.prev.After(context, task)
	}
}

// Handle method adapts the request object and creates the context
// to kickoff the execution of the handlers. After processing, the
// context and task is then adapted into the expected response for
// the caller.
func (r *Router) Handle(req interface{}) interface{} {
	context := r.createContext()
	task := r.adaptRequest(req)
	r.handler.Before(&context, &task)
	return r.adaptResponse(&task)
}

// NewRouter is a factory method to create a Router pointer with a
// default context creator, a given request and response adapter, and
// a list of handlers.
func NewRouter(requestAdapter RequestAdapter, responseAdapter ResponseAdapter,
	handlers ...Handler) *Router {
	return newRouter(DefaultContextCreator, requestAdapter, responseAdapter, handlers...)
}

// NewRouterWithContextCreator is a factory method to create a Router
// pointer with a given context creator, request and response adapter,
// and a list of handlers.
func NewRouterWithContextCreator(ctxCreator ContextCreator,
	requestAdapter RequestAdapter, responseAdapter ResponseAdapter, handlers ...Handler) *Router {
	return newRouter(ctxCreator, requestAdapter, responseAdapter, handlers...)
}

func newRouter(ctxCreator ContextCreator, requestAdapter RequestAdapter,
	responseAdapter ResponseAdapter, handlers ...Handler) *Router {
	return &Router{
		handler:       NewChainHandler(handlers...),
		createContext: ctxCreator,
		adaptRequest:  requestAdapter,
		adaptResponse: responseAdapter,
	}
}
