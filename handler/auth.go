package handler

import (
	"github.com/DukeRupert/haven/view/auth"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	db *pgxpool.Pool
}

func (h AuthHandler) HandleLogin(c echo.Context) error {
	return render(c, auth.Login())
}
