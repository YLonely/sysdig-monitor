package router

type localRoute struct {
	method  string
	path    string
	handler HandlerType
}

var _ Route = localRoute{}

func (l localRoute) Method() string {
	return l.method
}

func (l localRoute) Path() string {
	return l.path
}

func (l localRoute) Handler() HandlerType {
	return l.handler
}

func NewRoute(method, path string, handler HandlerType) Route {
	return localRoute{method: method, path: path, handler: handler}
}

func NewGetRoute(method, path string, handler HandlerType) Route {
	return NewRoute(MethodGet, path, handler)
}

func NewPostRoute(method, path string, handler HandlerType) Route {
	return NewRoute(MethodPost, path, handler)
}

func NewPutRoute(method, path string, handler HandlerType) Route {
	return NewRoute(MethodPut, path, handler)
}

func NewDeleteRoute(method, path string, handler HandlerType) Route {
	return NewRoute(MethodDelete, path, handler)
}
