package httpserver

import (
	echo "github.com/labstack/echo/v4"
)

type Router struct {}

func NewRouter() *Router { return &Router{} }

func (r *Router) Register(e *echo.Echo) {
	api := e.Group("/api")
	decisions := api.Group("/decisions")
	decisions.POST("/:model/:decision/evaluate", evaluateDecision)
}
