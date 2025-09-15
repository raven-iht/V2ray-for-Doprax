package main

import (
	"log"
	"net/http"
	"os"

	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"example.com/decision-backend/internal/http"
)

func main() {
	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.Use(middleware.Secure())
	e.GET("/healthz", func(c echo.Context) error { return c.String(http.StatusOK, "ok") })

	r := httpserver.NewRouter()
	r.Register(e)

	port := os.Getenv("PORT")
	if port == "" { port = "8080" }
	if err := e.Start(":" + port); err != nil { log.Fatal(err) }
}
