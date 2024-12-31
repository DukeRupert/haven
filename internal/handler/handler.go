package handler

import (
	"github.com/DukeRupert/haven/internal/mail"
	"github.com/DukeRupert/haven/internal/model/dto"
	"github.com/DukeRupert/haven/internal/repository"
	"github.com/DukeRupert/haven/web/view/page"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

// HandlerFunc is the type for your page handlers
type HandlerFunc func(c echo.Context, pageCtx *dto.PageContext) error

type Cfg struct {
	BaseURL string
}

type Config struct {
	Repos        *repository.Repositories
	Logger       zerolog.Logger
	BaseURL      string
	MailerConfig MailerConfig
}

type MailerConfig struct {
	BaseURL     string
	ServerToken string
	FromEmail   string
	FromName    string
}

type Handler struct {
	repos    *repository.Repositories
	logger   zerolog.Logger
	config   Cfg
	mailer   *mail.Mailer
	RouteCtx dto.RouteContext
}

func New(cfg Config) (*Handler, error) {
	// Initialize mail client
	mailClient := mail.NewClient(cfg.MailerConfig.ServerToken, cfg.Logger.With().Str("component", "postmark_client").Logger())
	mailer, err := mail.NewMailer(
		mailClient,
		cfg.MailerConfig.FromEmail,
		cfg.MailerConfig.FromName,
	)
	if err != nil {
		return nil, err
	}

	return &Handler{
		repos:  cfg.Repos,
		logger: cfg.Logger.With().Str("component", "handler").Logger(),
		config: Cfg{BaseURL: cfg.BaseURL},
		mailer: mailer,
	}, nil
}

func (h *Handler) GetLogin(c echo.Context) error {
	// If redirected from set-password
	// notify the user of the successful password update
	verify := c.QueryParam("verify") == "success"
	// Else simply return the login page
	return render(c, page.Login(verify))
}

func (h *Handler) GetHome(c echo.Context) error {
	return render(c, page.Landing())
}
