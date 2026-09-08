package main

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mzkki/ottodot-trial/internal/bootstrap"
	"github.com/mzkki/ottodot-trial/internal/config"
	"github.com/mzkki/ottodot-trial/internal/delivery/http/docs"
	"github.com/mzkki/ottodot-trial/internal/delivery/http/middleware"
	"github.com/mzkki/ottodot-trial/internal/delivery/http/routes"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
	"gorm.io/gorm"
)

type App struct {
	Config *viper.Viper
	DB     *gorm.DB
	Router *gin.Engine
	Logger *zap.Logger
}

// healthHandler returns system health status.
// @Summary      Health Check
// @Description  Returns system health status
// @Tags         System
// @Produce      json
// @Success      200 {object} map[string]string
// @Router       /health [get]
func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "ottodot-trial",
	})
}

func (a *App) initConfig() {
	v := viper.New()
	v.SetConfigFile(".env")
	v.SetConfigType("env")
	v.AutomaticEnv()
	if err := v.ReadInConfig(); err != nil && !os.IsNotExist(err) {
		// Log or handle non-exist errors if needed
	}

	// In test mode, route database queries to dedicated test database
	if os.Getenv("APP_ENV") == "test" {
		testDB := os.Getenv("DB_TEST_NAME")
		if testDB == "" {
			testDB = v.GetString("DB_TEST_NAME")
		}
		if testDB != "" {
			v.Set("DB_NAME", testDB)
		}
	}

	a.Config = v
}

func (a *App) initLogger() {
	log, err := config.NewLogger(a.Config)
	if err != nil {
		panic(fmt.Errorf("failed to init logger: %w", err))
	}
	a.Logger = log
}

func (a *App) initDatabase() {
	db, err := config.NewDatabase(a.Config)
	if err != nil {
		a.Logger.Fatal("failed_to_connect_database", zap.Error(err))
	}
	a.Logger.Info("database_connected")
	a.DB = db
}

// loadTemplates recursively parses all .html files from the templates directory.
func loadTemplates(dir string) (*template.Template, error) {
	funcMap := template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"divFloat": func(a, b int) float64 {
			if b == 0 {
				return 0
			}
			return float64(a) / float64(b)
		},
	}
	tmpl := template.New("").Funcs(funcMap)

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".html" {
			_, err := tmpl.ParseFiles(path)
			if err != nil {
				return fmt.Errorf("failed to parse template %s: %w", path, err)
			}
		}
		return nil
	})

	return tmpl, err
}

func (a *App) initRouter() {
	r := gin.New()
	r.RedirectTrailingSlash = true
	r.RedirectFixedPath = true

	// Load HTML templates
	tmpl, err := loadTemplates("templates")
	if err != nil {
		a.Logger.Fatal("failed_to_load_templates", zap.Error(err))
	}
	r.SetHTMLTemplate(tmpl)

	// Global middleware
	r.Use(middleware.RequestID())
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.RateLimit(middleware.NewIPRateLimiter(rate.Limit(50), 100)))
	r.Use(middleware.AccessLog(a.Logger))
	r.Use(middleware.RecoveryLog(a.Logger))

	// Health check
	r.GET("/api/health", healthHandler)
	r.GET("/api/v1/health", healthHandler)

	// API Docs (Scalar UI + OpenAPI Spec)
	docsGroup := r.Group("/api/v1/docs")
	{
		docsGroup.GET("", docs.DocsHandler)
		docsGroup.GET("/", docs.DocsHandler)
		docsGroup.StaticFile("/openapi.json", "./docs/swagger.json")
	}
	r.GET("/docs", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/api/v1/docs/")
	})

	// Initialize handlers via Wire DI
	handlers, err := bootstrap.InitializeHandlers(a.DB, a.Config)
	if err != nil {
		a.Logger.Fatal("failed_to_initialize_handlers", zap.Error(err))
	}

	// Register all routes
	routes.RegisterRoutes(r, handlers)

	a.Router = r
}

func (a *App) Run() {
	defer a.Logger.Sync()

	port := a.Config.GetString("APP_PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           a.Router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	a.Logger.Info("server_starting", zap.String("addr", port))

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.Logger.Fatal("server_listen_failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	a.Logger.Info("server_shutting_down")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		a.Logger.Fatal("server_shutdown_failed", zap.Error(err))
	}

	if a.DB != nil {
		if sqlDB, err := a.DB.DB(); err == nil {
			if closeErr := sqlDB.Close(); closeErr != nil {
				a.Logger.Warn("failed_to_close_db", zap.Error(closeErr))
			} else {
				a.Logger.Info("database_connections_closed")
			}
		}
	}

	a.Logger.Info("server_stopped")
}

func newApp() *App {
	app := &App{}
	app.initConfig()
	app.initLogger()
	app.initDatabase()
	app.initRouter()
	return app
}
