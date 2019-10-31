package router

import (
	"github.com/gin-gonic/gin"
)

type HandlerType func(c *gin.Context)

// Router defines an interface to specify a group of routes ot add to the server
type Router interface {
	Routes() []Route
}

// Route defines an individual API route in the sysdig-monitor server
type Route interface {
	Method() string
	Path() string
	Handler() HandlerType
}

const (
	MethodPost   = "POST"
	MethodGet    = "GET"
	MethodPut    = "PUT"
	MethodDelete = "DELETE"
	MethodPatch  = "PATCH"
)
