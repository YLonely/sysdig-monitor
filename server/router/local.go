package router

type router struct {
	routes []Route
}

var _ Router = &router{}

func NewRouter() Router {
	return &router{}
}

func (r *router) Routes() []Route {
	return r.routes
}

func (r *router) AddRoute(path, method string, handler HandlerType) {
	r.routes = append(r.routes, route{method: method, path: path, handler: handler})
}

type groupRouter struct {
	router
	prefix string
}

var _ Router = &groupRouter{}

func NewGroupRouter(prefix string) Router {
	return &groupRouter{prefix: prefix}
}

func (gr *groupRouter) AddRoute(path, method string, handler HandlerType) {
	gr.routes = append(gr.routes, route{path: gr.prefix + path, method: method, handler: handler})
}

type route struct {
	method  string
	path    string
	handler HandlerType
}

var _ Route = &route{}

func (l route) Method() string {
	return l.method
}

func (l route) Path() string {
	return l.path
}

func (l route) Handler() HandlerType {
	return l.handler
}
