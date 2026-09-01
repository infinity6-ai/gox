package httpzclient

// Handler defines the function signature for a client-side request handler.
// It takes a request and returns a response and an error.
type Handler func(req *Req) (*Resp, error)

// Filter is a middleware function that intercepts a request, allowing for modification
// of the request or response. It receives the current request and the next handler in the chain.
type Filter func(req *Req, next Handler) (*Resp, error)
