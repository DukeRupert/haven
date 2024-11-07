package main

import (
	"net/http"

	"github.com/DukeRupert/haven/handler"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	userHandler := handler.UserHandler{}
	e.GET("/user", userHandler.HandleUserShow)
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World! Welcome to Haven.")
	})
	e.Logger.Fatal(e.Start(":1323"))
}
