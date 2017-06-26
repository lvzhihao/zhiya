package goutils

import (
	"context"
	"html/template"
	"io"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/tylerb/graceful"
)

type EchoHost struct {
	Echo *echo.Echo
}

func NewEcho() *echo.Echo {
	app := echo.New()
	app.Use(middleware.Logger())
	app.Use(middleware.Recover())
	app.Use(middleware.Secure())
	app.Use(middleware.RequestID())
	return app
}

func NewEchoHosts(hosts map[string]EchoHost) *echo.Echo {
	app := echo.New()
	app.Any("/*", func(c echo.Context) (err error) {
		req := c.Request()
		if host, ok := hosts[req.Host]; !ok {
			err = echo.ErrNotFound
		} else {
			host.Echo.ServeHTTP(c.Response(), req)
		}
		return
	})
	return app
}

func EchoStartWithGracefulShutdown(app *echo.Echo, addr string) {
	if CompareRuntimeVersion("go1.8") {
		EchoStartWithGracefulShutdownThanGo18(app, addr)
	} else {
		EchoStartWithGracefulShutdownUseThirdPlugin(app, addr)
	}
}

func EchoStartWithGracefulShutdownThanGo18(app *echo.Echo, addr string) {
	// Start server
	go func() {
		if err := app.Start(addr); err != nil {
			app.Logger.Info("shutting down the server")
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := app.Shutdown(ctx); err != nil {
		app.Logger.Fatal(err)
	}
}

func EchoStartWithGracefulShutdownUseThirdPlugin(app *echo.Echo, addr string) {
	app.Server.Addr = addr
	app.Logger.Info(graceful.ListenAndServe(app.Server, 5*time.Second))
}

/*
 * echo templater support
 */
type EchoTemplate struct {
	templates *template.Template
}

func NewEchoRenderer(name, patter string) *EchoTemplate {
	return &EchoTemplate{
		templates: NewTemplate(name, patter),
	}
}

func (t *EchoTemplate) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}
