package routes

import (
	"fmt"
	"time"

	"llm-backend/internal/config"
	"llm-backend/internal/database"
	"llm-backend/internal/handlers"
	"llm-backend/internal/middleware"
	"llm-backend/internal/models"
	"llm-backend/internal/services"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, db *database.DB, cfg *config.Config) {
	// 初始化监控和日志系统
	metricsCollector := services.GetGlobalMetricsCollector()
	logManager := services.GetGlobalLogManager()

	// 添加全局中间件
	r.Use(middleware.GlobalRateLimit())            // 全局限流
	r.Use(metricsCollector.MonitoringMiddleware()) // 监控中间件
	r.Use(logManager.LoggingMiddleware())          // 日志中间件

	// 初始化仓库
	userRepo := models.NewUserRepository(db)
	apiCallRepo := models.NewAPICallRepository(db)

	// 初始化模型管理器
	modelManager, err := services.NewModelManager(cfg)
	if err != nil {
		panic("初始化模型管理器失败: " + err.Error())
	}

	// 初始化服务
	llmService := services.NewLLMService("", modelManager)

	// 初始化服务发现和负载均衡
	serviceRegistry := modelManager.GetServiceRegistry()
	loadBalancer := services.NewLoadBalancer(serviceRegistry, "round_robin")
	loadBalancer.StartCleanupRoutine()

	// 初始化集群管理
	nodeID := fmt.Sprintf("node-%d", time.Now().Unix())
	clusterManager := services.NewClusterManager(nodeID, "127.0.0.1", 8080, serviceRegistry)

	// 初始化处理器
	authHandler := handlers.NewAuthHandler(userRepo, cfg)
	llmHandler := handlers.NewLLMHandler(userRepo, apiCallRepo, llmService)
	modelHandler := handlers.NewModelHandler(modelManager)
	gatewayHandler := handlers.NewGatewayHandler(modelManager, llmService)
	serviceDiscoveryHandler := services.NewServiceDiscoveryHandler(serviceRegistry, loadBalancer)
	monitoringHandler := services.NewMonitoringHandler(metricsCollector)
	logHandler := services.NewLogHandler(logManager)
	clusterHandler := services.NewClusterHandler(clusterManager)
	webHandler := handlers.NewWebHandler()

	// 初始化任务服务和处理器
	taskService := services.NewTaskService(modelManager)
	taskHandler := handlers.NewTaskHandler(taskService, userRepo)

	// Web 页面
	r.GET("/", webHandler.Index)

	// 健康检查
	r.GET("/health", gatewayHandler.HealthCheck)
	r.GET("/status", gatewayHandler.SystemStatus)

	// API路由组
	api := r.Group("/api/v1")
	{
		// 认证路由（无需token）
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// 需要认证的路由
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		protected.Use(middleware.UserRateLimit()) // 用户限流
		{
			// 用户相关
			protected.GET("/auth/profile", authHandler.GetProfile)

			// LLM相关
			protected.POST("/chat", llmHandler.Chat)
			protected.GET("/history", llmHandler.GetHistory)
			protected.GET("/stats", llmHandler.GetStats)

			// 模型管理相关
			models := protected.Group("/models")
			models.Use(middleware.ModelRateLimit(10, 5)) // 模型限流：每秒5个请求，桶容量10
			{
				models.GET("/", modelHandler.GetAvailableModels)         // 获取可用模型列表
				models.GET("/running", modelHandler.GetRunningModels)    // 获取运行中的模型
				models.GET("/metrics", modelHandler.GetModelMetrics)     // 获取模型性能指标
				models.POST("/:name/start", modelHandler.StartModel)     // 启动模型
				models.POST("/:name/stop", modelHandler.StopModel)       // 停止模型
				models.POST("/:name/restart", modelHandler.RestartModel) // 重启模型
				models.GET("/:name/status", modelHandler.GetModelStatus) // 获取模型状态
				models.POST("/:name/chat", modelHandler.ChatWithModel)   // 与指定模型对话
			}

			// API 网关路由（OpenAI 兼容）
			v1 := protected.Group("/v1")
			{
				// OpenAI 兼容接口
				v1.POST("/chat/completions", gatewayHandler.ChatCompletions)
				v1.GET("/models", gatewayHandler.ListModels)
				v1.POST("/batch", gatewayHandler.BatchRequest)

				// 直接代理到 llama.cpp 服务器
				v1.Any("/proxy/:model/*path", gatewayHandler.ProxyToLlamaCpp)
			}

			// 服务发现和负载均衡
			discovery := protected.Group("/discovery")
			{
				discovery.POST("/register", serviceDiscoveryHandler.RegisterService)
				discovery.DELETE("/:service/:instance", serviceDiscoveryHandler.DeregisterService)
				discovery.GET("/services", serviceDiscoveryHandler.DiscoverServices)
				discovery.GET("/services/:service", serviceDiscoveryHandler.DiscoverServices)
				discovery.GET("/stats", serviceDiscoveryHandler.GetServiceStats)
				discovery.GET("/load-balancer/strategy", serviceDiscoveryHandler.GetLoadBalancingStrategy)
				discovery.PUT("/load-balancer/strategy", serviceDiscoveryHandler.SetLoadBalancingStrategy)
			}

			// 监控相关
			monitoring := protected.Group("/monitoring")
			{
				monitoring.GET("/metrics", monitoringHandler.GetMetrics)
				monitoring.GET("/metrics/:name", monitoringHandler.GetMetricsByName)
				monitoring.GET("/system", monitoringHandler.GetSystemStats)
				monitoring.POST("/metrics", monitoringHandler.RecordCustomMetric)
			}

			// 日志相关
			logs := protected.Group("/logs")
			{
				logs.GET("/", logHandler.GetLogs)
				logs.GET("/loggers", logHandler.GetLoggers)
				logs.POST("/", logHandler.WriteLog)
			}

			// 集群管理
			cluster := protected.Group("/cluster")
			{
				cluster.POST("/join", clusterHandler.JoinCluster)
				cluster.POST("/leave", clusterHandler.LeaveCluster)
				cluster.POST("/heartbeat", clusterHandler.Heartbeat)
				cluster.GET("/nodes", clusterHandler.GetNodes)
				cluster.GET("/stats", clusterHandler.GetClusterStats)
				cluster.GET("/select", clusterHandler.SelectNode)
			}

			// 限流状态查询
			protected.GET("/rate-limit/status", middleware.GetRateLimitStatus())

			// 任务处理接口
			tasks := protected.Group("/tasks")
			{
				tasks.POST("/convert", taskHandler.ConvertFileFormat)
				tasks.POST("/homework", taskHandler.GradeHomework)
				tasks.POST("/subtitle", taskHandler.ProcessSubtitle)
				tasks.GET("/formats", taskHandler.GetSupportedFormats)
				tasks.GET("/status/:id", taskHandler.GetTaskStatus)
				tasks.POST("/upload", taskHandler.UploadFile)
			}
		}
	}

	// 静态文件服务（如果需要）
	r.Static("/static", "./static")
}
