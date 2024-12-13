// internal/handler/handler.go
package handler

import (
	"github.com/DukeRupert/haven/internal/auth"
	"github.com/DukeRupert/haven/internal/model/dto"
	"github.com/DukeRupert/haven/internal/repository"
	"github.com/DukeRupert/haven/web/view/page"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

// HandlerFunc is the type for your page handlers
type HandlerFunc func(c echo.Context, pageCtx *dto.PageContext) error

type Config struct {
	Repos  *repository.Repositories
	Auth   *auth.Middleware
	Logger zerolog.Logger
}

type Handler struct {
	repos    *repository.Repositories
	auth     *auth.Middleware
	logger   zerolog.Logger
	RouteCtx dto.RouteContext
}

func New(cfg Config) *Handler {
	return &Handler{
		repos:  cfg.Repos,
		auth:   cfg.Auth,
		logger: cfg.Logger.With().Str("component", "handler").Logger(),
	}
}

func (h *Handler) GetLogin(c echo.Context) error {
	return render(c, page.Login())
}

func (h *Handler) GetHome(c echo.Context) error {
	return render(c, page.Landing())
}
