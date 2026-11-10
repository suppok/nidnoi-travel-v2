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
	"backend/pkg/modules/shared/client/deployhook"
	"backend/pkg/modules/shared/client/line"
	"backend/pkg/modules/shared/client/objectstorage"
	"backend/pkg/modules/shared/services/password"
	"backend/pkg/modules/shared/services/revalidation"
	"backend/pkg/registry"

	airlineusecase "backend/pkg/modules/airline/usecase"
	authusecase "backend/pkg/modules/auth/usecase"
	bookingusecase "backend/pkg/modules/booking/usecase"
	linebookingusecase "backend/pkg/modules/linebooking/usecase"
	linecontactusecase "backend/pkg/modules/linecontact/usecase"
	cityusecase "backend/pkg/modules/city/usecase"
	countryusecase "backend/pkg/modules/country/usecase"
	mediausecase "backend/pkg/modules/media/usecase"
	siteusecase "backend/pkg/modules/site/usecase"
	tourprogramusecase "backend/pkg/modules/tourprogram/usecase"
	periodusecase "backend/pkg/modules/tourprogramperiod/usecase"
	userusecase "backend/pkg/modules/user/usecase"
	wholesaleusecase "backend/pkg/modules/wholesale/usecase"
)

const (
	configPath      = "config/config_api.yml"
	shutdownTimeout = 10 * time.Second
)

type App struct {
	fiber       *fiber.App
	config      *config.Config
	logger      *zap.Logger
	db          *bun.DB
	revalidator revalidation.Worker
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
	deployHook := deployhook.New(cfg.DeployHookURL)
	lineClient := line.New(cfg.LINE)
	revalidator := revalidation.New(deployHook, revalidation.NewRepositoryBuildLog(repositoryRegistry.SiteBuildRepository, log), log, cfg.Revalidation)
	shareRegistry := &registry.ShareRegistry{
		ServiceRegistry: registry.ServiceRegistry{Password: passwords, Revalidator: revalidator},
		ClientRegistry:  registry.ClientRegistry{ObjectStorage: objectstorage.New(cfg.R2), DeployHook: deployHook, LINE: lineClient},
	}
	if !cfg.R2.Configured() {
		log.Warn("R2 is not configured: media uploads return STORAGE_UNAVAILABLE until every R2_* variable is set")
	}
	if !deployHook.Configured() {
		log.Warn("CLOUDFLARE_DEPLOY_HOOK_URL is not set: customer-site rebuilds are logged and skipped")
	}
	if !lineClient.Configured() {
		log.Warn("LINE is not configured: LIFF endpoints return LINE_NOT_CONFIGURED until LINE_LIFF_CHANNEL_ID is set")
	}
	if !lineClient.WebhookConfigured() {
		log.Warn("LINE_CHANNEL_SECRET is not set: the LINE webhook returns LINE_NOT_CONFIGURED")
	}

	// Use cases
	authUseCase := authusecase.NewAuthUseCase(coreRegistry, shareRegistry, repositoryRegistry)
	userUseCase := userusecase.NewUserUseCase(coreRegistry, shareRegistry, repositoryRegistry)
	countryUseCase := countryusecase.NewCountryUseCase(coreRegistry, shareRegistry, repositoryRegistry)
	cityUseCase := cityusecase.NewCityUseCase(coreRegistry, shareRegistry, repositoryRegistry)
	wholesaleUseCase := wholesaleusecase.NewWholesaleUseCase(coreRegistry, shareRegistry, repositoryRegistry)
	airlineUseCase := airlineusecase.NewAirlineUseCase(coreRegistry, shareRegistry, repositoryRegistry)
	mediaUseCase := mediausecase.NewMediaUseCase(coreRegistry, shareRegistry, repositoryRegistry)
	tourProgramUseCase := tourprogramusecase.NewTourProgramUseCase(coreRegistry, shareRegistry, repositoryRegistry)
	periodUseCase := periodusecase.NewTourProgramPeriodUseCase(coreRegistry, shareRegistry, repositoryRegistry)
	siteUseCase := siteusecase.NewSiteUseCase(coreRegistry, shareRegistry, repositoryRegistry)
	bookingUseCase := bookingusecase.NewBookingUseCase(coreRegistry, shareRegistry, repositoryRegistry)
	lineBookingUseCase := linebookingusecase.NewLineBookingUseCase(coreRegistry, shareRegistry, repositoryRegistry)
	lineContactUseCase := linecontactusecase.NewLineContactUseCase(coreRegistry, shareRegistry, repositoryRegistry)

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
	liffAuth := middleware.NewLIFFAuth(authUseCase)

	v1 := f.Group("/api/v1")
	handler.NewHealthAPIHandler(v1, apiHandler).Init()
	handler.NewAuthAPIHandler(v1, apiHandler, cfg, authUseCase).Init(jwtAuth, liffAuth, authPermissions)
	handler.NewUserAPIHandler(v1, apiHandler, userUseCase).Init(jwtAuth, authPermissions)
	handler.NewCountryAPIHandler(v1, apiHandler, countryUseCase).Init(jwtAuth, authPermissions)
	handler.NewCityAPIHandler(v1, apiHandler, cityUseCase).Init(jwtAuth, authPermissions)
	handler.NewWholesaleAPIHandler(v1, apiHandler, wholesaleUseCase).Init(jwtAuth, authPermissions)
	handler.NewAirlineAPIHandler(v1, apiHandler, airlineUseCase).Init(jwtAuth, authPermissions)
	handler.NewMediaAPIHandler(v1, apiHandler, mediaUseCase, cfg.Media.MaxUploadMB).Init(jwtAuth, authPermissions)
	handler.NewTourProgramAPIHandler(v1, apiHandler, tourProgramUseCase).Init(jwtAuth, authPermissions)
	handler.NewTourProgramPeriodAPIHandler(v1, apiHandler, periodUseCase).Init(jwtAuth, authPermissions)
	handler.NewSiteAPIHandler(v1, apiHandler, siteUseCase).Init(jwtAuth, authPermissions)
	handler.NewBookingAPIHandler(v1, apiHandler, cfg, bookingUseCase).Init(jwtAuth, authPermissions)
	handler.NewLineBookingAPIHandler(v1, apiHandler, lineBookingUseCase, mediaUseCase, cfg.Media.MaxUploadMB).Init(jwtAuth, liffAuth, authPermissions)
	handler.NewLineContactAPIHandler(v1, apiHandler, lineContactUseCase).Init(jwtAuth, liffAuth, authPermissions)

	return &App{fiber: f, config: cfg, logger: log, db: db, revalidator: revalidator}, nil
}

// Run serves until ctx is cancelled, then shuts down gracefully.
func (a *App) Run(ctx context.Context) error {
	defer func() {
		_ = a.db.Close()
		_ = a.logger.Sync()
	}()

	// The rebuild worker stops after the HTTP server, so edits made during shutdown are still flushed.
	workerCtx, stopWorker := context.WithCancel(context.Background())
	workerDone := make(chan struct{})
	go func() { a.revalidator.Run(workerCtx); close(workerDone) }()
	defer func() {
		stopWorker()
		<-workerDone
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
