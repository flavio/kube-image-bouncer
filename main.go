package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"

	"github.com/flavio/kube-image-bouncer/handlers"
)

func main() {
	var cert, key string
	var port int

	flag.IntVar(&port, "port", 1323, "Port number to listen to")
	flag.StringVar(&cert, "cert", "server.cert", "TLS certificate file")
	flag.StringVar(&key, "key", "server.key", "TLS key")

	flag.Parse()

	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.POST("/policy", handlers.PostImagePolicy())

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}\n",
	}))

	e.Logger.SetLevel(log.DEBUG)

	e.Logger.Fatal(
		e.StartTLS(
			fmt.Sprintf(":%d", port),
			cert,
			key))
}
