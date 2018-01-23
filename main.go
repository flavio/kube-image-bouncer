package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
	"gopkg.in/urfave/cli.v1"

	"github.com/flavio/kube-image-bouncer/handlers"
)

func main() {
	var cert, key string
	var port int
	var debug bool

	app := cli.NewApp()
	app.Name = "kube-image-bouncer"
	app.Usage = "webhook endpoint for kube image policy admission controller"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "cert, c",
			Value:       "server.cert",
			Usage:       "Path to the certificate to use",
			EnvVar:      "BOUNCER_CERTIFICATE",
			Destination: &cert,
		},
		cli.StringFlag{
			Name:        "key, k",
			Value:       "server.key",
			Usage:       "Path to the key to use",
			EnvVar:      "BOUNCER_KEY",
			Destination: &key,
		},
		cli.IntFlag{
			Name:        "port, p",
			Value:       1323,
			Usage:       "`PORT` to listen to",
			EnvVar:      "BOUNCER_PORT",
			Destination: &port,
		},
		cli.BoolFlag{
			Name:        "debug",
			Usage:       "Enable extra debugging",
			EnvVar:      "BOUNCER_DEBUG",
			Destination: &debug,
		},
	}

	app.Action = func(c *cli.Context) error {
		e := echo.New()
		e.GET("/", func(c echo.Context) error {
			return c.String(http.StatusOK, "Hello, World!")
		})
		e.POST("/policy", handlers.PostImagePolicy())

		e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
			Format: "method=${method}, uri=${uri}, status=${status}\n",
		}))

		if debug {
			e.Logger.SetLevel(log.DEBUG)
		}

		err := e.StartTLS(fmt.Sprintf(":%d", port), cert, key)
		if err != nil {
			return cli.NewExitError(err, 1)
		}

		return nil
	}

	app.Run(os.Args)
}
