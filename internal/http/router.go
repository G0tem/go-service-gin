package http_router

import (
	"time"

	"github.com/G0tem/go-service-gin/internal/domain/ports"
	handler "github.com/G0tem/go-service-gin/internal/http/handlers"
	"github.com/G0tem/go-service-gin/internal/http/middleware"
	"github.com/gin-gonic/gin"
	promclient "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func NewRouter(oh *handler.OrderHandler, ah *handler.AuthHandler, tm ports.TokenManager, gatherer promclient.Gatherer) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery(), middleware.Timeout(10*time.Second), otelgin.Middleware("order-service"))

	v1 := r.Group("/api/v1")
	v1.GET("/healthz", func(c *gin.Context) { c.String(200, "ok") })

	v1.POST("/auth/login", ah.Login)

	// 📈 Prometheus endpoint с явным Gatherer
	r.GET("/metrics", gin.WrapH(promhttp.HandlerFor(gatherer, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})))

	// 📚 Swagger UI (доступен на /swagger/index.html)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler,
		ginSwagger.URL("/swagger/doc.json"),     // URL для загрузки swagger.json
		ginSwagger.DefaultModelsExpandDepth(-1), // скрыть модели по умолчанию
	))

	// 🔐 Защищённые маршруты с RBAC
	orders := v1.Group("/orders")
	orders.Use(middleware.JWTAuth(tm)) // 1. Расшифровка токена

	orders.POST("",
		middleware.RequireScope("orders:write"), // 2. Проверка скоупа
		oh.Create,
	)

	// orders.GET("",
	// 	middleware.RequireRole("user", "admin"), // 3. Проверка роли
	// 	oh.List,
	// )

	// orders.DELETE("/:id",
	// 	middleware.RequireRole("admin"), // Комбинация
	// 	middleware.RequireScope("orders:delete"),
	// 	oh.Delete,
	// )

	return r
}
