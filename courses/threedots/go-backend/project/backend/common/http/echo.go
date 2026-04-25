package http

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"

	"eats/backend/common"
)

func NewEcho() *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	useMiddlewares(e)
	e.HTTPErrorHandler = common.EchoErrorHandler
	e.Logger = common.NewEchoSlogAdapter(slog.Default())

	e.GET("/health", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	return e
}
