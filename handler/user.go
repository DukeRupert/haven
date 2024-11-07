package handler

import (
	"github.com/DukeRupert/haven/view/user"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

type UserHandler struct{
    db *pgxpool.Pool
}

func (h UserHandler) HandleUserShow(c echo.Context) error {
	return render(c, user.Show())
}
