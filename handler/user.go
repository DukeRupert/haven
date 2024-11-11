package handler

import (
	"github.com/DukeRupert/haven/view/auth"
	"github.com/DukeRupert/haven/view/user"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	db *pgxpool.Pool
}

// NewUserHandler creates a new handler with both pool and store
func NewUserHandler(pool *pgxpool.Pool) *UserHandler {
	return &UserHandler{
		db: pool,
	}
}

func (h *UserHandler) GetLogin(c echo.Context) error {
	return render(c, auth.Login())
}

func (h UserHandler) HandleUserShow(c echo.Context) error {
	return render(c, user.Show())
}
