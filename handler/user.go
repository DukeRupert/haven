package handler

import (
	"github.com/DukeRupert/haven/db"
	"github.com/DukeRupert/haven/view/auth"
	"github.com/DukeRupert/haven/view/user"
	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	db *db.DB
}

// NewUserHandler creates a new handler with both pool and store
func NewUserHandler(db *db.DB) *UserHandler {
	return &UserHandler{
		db: db,
	}
}

func (h *UserHandler) GetLogin(c echo.Context) error {
	return render(c, auth.Login())
}

func (h UserHandler) HandleUserShow(c echo.Context) error {
	return render(c, user.Show())
}
