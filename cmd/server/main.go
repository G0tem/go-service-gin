package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/G0tem/go-service-gin/docs"
	"github.com/G0tem/go-service-gin/internal/application"
	http_router "github.com/G0tem/go-service-gin/internal/delivery/http"
	handler "github.com/G0tem/go-service-gin/internal/delivery/http/handlers"
	"github.com/G0tem/go-service-gin/internal/infrastructure/jwt"
	"github.com/G0tem/go-service-gin/internal/infrastructure/postgres"
	infra_product "github.com/G0tem/go-service-gin/internal/infrastructure/postgres/product"
	"github.com/G0tem/go-service-gin/internal/infrastructure/rabbitmq"
	"github.com/G0tem/go-service-gin/internal/otel"
	"github.com/G0tem/go-service-gin/pkg/config"
)

// @title Order Service API
// @version 1.0
// @description Microservice for order management with JWT auth, Postgres, Redis, RabbitMQ
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer {token}" to authenticate
func main() {
	// 1. Загружаем конфиг (с валидацией!)
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ failed to load config: %v", err)
	}

	ctx := context.Background()

	// 2. Инициализация трейсинга
	shutdownTracer, err := otel.InitTracing(ctx, cfg.ServiceName, cfg.OTelTrace)
	exitOnError(err, "otel tracing")
	defer shutdownTracer(ctx)

	// 3. Инициализация метрик
	gatherer, metrics, meterProvider, err := otel.InitMetrics(ctx, cfg.ServiceName, cfg.OTelMetric)
	exitOnError(err, "otel metrics")
	if meterProvider != nil {
		defer func() {
			if err := meterProvider.Shutdown(ctx); err != nil {
				log.Printf("⚠️ failed to shutdown meter provider: %v", err)
			}
		}()
	}

	// Инфраструктура
	pool, err := postgres.NewPool(ctx, cfg.PostgresDSN)
	exitOnError(err, "postgres")
	defer pool.Close()

	rmq, err := rabbitmq.NewClient(ctx, cfg.RabbitMQURL)
	exitOnError(err, "rabbitmq")
	defer rmq.Close()

	// Domain & UseCases
	orderRepo := postgres.NewOrderRepo(pool)
	productRepo := infra_product.NewProductRepo(pool)
	tm := jwt.NewManager(cfg.JWTSecret, cfg.AccessTokenTTL)
	authHandler := handler.NewAuthHandler(tm)
	createOrderUC := application.NewCreateOrderHandler(orderRepo, rmq, metrics)
	orderHandler := handler.NewOrderHandler(createOrderUC)

	router := http_router.NewRouter(orderHandler, authHandler, tm, gatherer, productRepo)

	srv := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Printf("🚀 HTTP server listening on %s", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("⏳ shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("✅ server stopped")
}

func exitOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("failed to init %s: %v", msg, err)
	}
}
