package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	"github.com/uptrace/bun"
	"go.uber.org/zap"

	"backend/cmd/api/handler"
	_ "backend/docs" // swagger spec
	"backend/pkg/app/initializer"
	"backend/pkg/core/api"
	"backend/pkg/core/api/middleware"
	"backend/pkg/core/auth/jwt"
	"backend/pkg/core/auth/permission"
	"backend/pkg/core/config"
	"backend/pkg/core/database"
	"backend/pkg/core/database/query"
	corelogger "backend/pkg/core/logger"
	"backend/pkg/modules/shared/services/password"
	"backend/pkg/registry"

	authusecase "backend/pkg/modules/auth/usecase"
	userusecase "backend/pkg/modules/user/usecase"
)

const (
	configPath      = "config/config_api.yml"
	shutdownTimeout = 10 * time.Second
)

type App struct {
	fiber  *fiber.App
	config *config.Config
	logger *zap.Logger
	db     *bun.DB
}

// New wires config → infrastructure → registries → use cases → handlers.
func New(ctx context.Context) (*App, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, err
	}

	log, err := corelogger.New(cfg.AppEnv)
	if err != nil {
		return nil, fmt.Errorf("logger: %w", err)
	}

	db, err := database.Open(ctx, cfg)
	if err != nil {
		return nil, err
	}

	permissions, err := permission.New(cfg.Permission.ModelPath, cfg.Permission.PolicyPath)
	if err != nil {
		return nil, err
	}

	passwords, err := password.NewService(cfg.Auth.BcryptCost)
	if err != nil {
		return nil, err
	}

	coreRegistry := &registry.CoreRegistry{
		Config:      cfg,
		Logger:      log,
		AppDB:       db,
		DBTx:        query.NewDBTx(db),
		JWT:         jwt.New(cfg.JWTSecret, cfg.Auth.AccessTokenTTL),
		Permissions: permissions,
	}
	repositoryRegistry := initializer.NewRepositoryRegistry(db)
	shareRegistry := &registry.ShareRegistry{
		ServiceRegistry: registry.ServiceRegistry{Password: passwords},
	}

	// Use cases
	authUseCase := authusecase.NewAuthUseCase(coreRegistry, shareRegistry, repositoryRegistry)
	userUseCase := userusecase.NewUserUseCase(coreRegistry, shareRegistry, repositoryRegistry)

	// HTTP
	response := api.NewApiResponseHandler(log)
	f := fiber.New(fiber.Config{
		BodyLimit:    cfg.Server.BodyLimitMB * 1024 * 1024,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		ErrorHandler: response.FiberErrorHandler,
	})
	f.Use(
		cors.New(cors.Config{
			AllowOrigins:     strings.Join(cfg.CORSAllowedOrigins, ","),
			AllowMethods:     "GET,POST,PUT,OPTIONS",
			AllowHeaders:     "Authorization,Content-Type",
			AllowCredentials: true, // refresh cookie
			MaxAge:           600,
		}),
		logger.New(),
		recover.New(),
	)

	if !cfg.IsPrd() {
		f.Get("/swagger/*", swagger.HandlerDefault)
	}

	apiHandler := api.NewApiHandler(response, cfg.Server.ClientIPHeader)
	jwtAuth := middleware.NewAuth(coreRegistry.JWT, authUseCase)
	authPermissions := middleware.NewAuthPermissions(permissions)

	v1 := f.Group("/api/v1")
	handler.NewHealthAPIHandler(v1, apiHandler).Init()
	handler.NewAuthAPIHandler(v1, apiHandler, cfg, authUseCase).Init(jwtAuth)
	handler.NewUserAPIHandler(v1, apiHandler, userUseCase).Init(jwtAuth, authPermissions)

	return &App{fiber: f, config: cfg, logger: log, db: db}, nil
}

// Run serves until ctx is cancelled, then shuts down gracefully.
func (a *App) Run(ctx context.Context) error {
	defer func() {
		_ = a.db.Close()
		_ = a.logger.Sync()
	}()

	errCh := make(chan error, 1)
	go func() { errCh <- a.fiber.Listen(":" + a.config.Port) }()
	a.logger.Info("api started", zap.String("env", a.config.AppEnv), zap.String("port", a.config.Port))

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return a.fiber.ShutdownWithTimeout(shutdownTimeout)
	}
}
