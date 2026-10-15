package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"skill-credential-backend/chain"
	"skill-credential-backend/config"
	"skill-credential-backend/handlers"
	"skill-credential-backend/middleware"
	"skill-credential-backend/models"
)

func main() {
	cfg := config.Load()

	// ---------- 1. 连接 MySQL ----------
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("连接 MySQL 失败: %v\n请确认 MySQL 已启动并执行 sql/init.sql", err)
	}

	// 自动建表（幂等）
	if err := db.AutoMigrate(
		&models.User{},
		&models.Achievement{},
		&models.Credential{},
		&models.VerifyLog{},
		&models.SysLog{},
	); err != nil {
		log.Fatalf("建表失败: %v", err)
	}
	log.Println("[DB] MySQL 连接成功，表结构就绪")

	// ---------- 2. 连接联盟链 ----------
	var chainClient *chain.ChainClient
	if cfg.ContractAddr != "" {
		chainClient, err = chain.NewChainClient(cfg)
		if err != nil {
			log.Printf("[CHAIN] 警告: 链连接失败(服务仍可启动, 上链接口暂不可用): %v", err)
		} else {
			log.Printf("[CHAIN] 联盟链连接成功 rpc=%s contract=%s", cfg.ChainRPC, cfg.ContractAddr)
		}
	} else {
		log.Println("[CHAIN] 未配置 CONTRACT_ADDR，上链接口暂不可用（部署合约后填写）")
	}

	// 确保上传目录存在
	os.MkdirAll(cfg.UploadDir, 0755)

	// ---------- 3. 初始化处理器 ----------
	authH := handlers.NewAuthHandler(db, cfg, chainClient)
	achH := handlers.NewAchievementHandler(db, cfg, chainClient)
	credH := handlers.NewCredentialHandler(db, cfg, chainClient)
	verifyH := handlers.NewVerifyHandler(db, cfg, chainClient)
	logH := handlers.NewLogHandler(db, cfg)
	uploadH := handlers.NewUploadHandler(db, cfg, chainClient)
	qrH := handlers.NewQRCodeHandler(db, cfg)

	// 启动日志哈希批量上链协程（论文 5.3.7：定期批量上链）
	handlers.StartLogOnChainWorker(db, chainClient, 30)

	// ---------- 4. 路由 ----------
	r := gin.Default()

	// 跨域（开发环境允许前端本地端口）
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	api := r.Group("/api")
	{
		// 公开接口
		api.POST("/auth/register", authH.Register)
		api.POST("/auth/login", authH.Login)

		// 外部核验（允许未登录游客核验，但做 IP 限流防刷链上交易）
		api.POST("/verify",
			middleware.RateLimitPerIP(10, time.Minute),
			verifyH.Verify)

		// 需登录
		authed := api.Group("", middleware.AuthRequired(cfg.JWTSecret))
		{
			authed.GET("/user/profile", authH.Profile)
			authed.GET("/user/lookup-student",
				middleware.RoleRequired("school_admin", "company"),
				authH.LookupStudent)

			// 成果录入：院校管理员 + 实训企业
			authed.POST("/achievement",
				middleware.RoleRequired("school_admin", "company"),
				achH.Create)
			authed.GET("/achievement", achH.List)

			// 文件上传
			authed.POST("/upload", uploadH.Upload)

			// 上链存证：院校管理员 + 实训企业
			authed.POST("/credential/store",
				middleware.RoleRequired("school_admin", "company"),
				credH.Store)

			// 学生查本人存证 / 管理端查全部
			authed.GET("/credential/my", credH.MyCredentials)

			// 检索
			authed.GET("/credential/search",
				middleware.RoleRequired("school_admin", "company"),
				credH.Search)

			// 核验历史：仅招聘企业（本人）与院校管理员（审计）可见
			authed.GET("/verify/history",
				middleware.RoleRequired("employer", "school_admin"),
				verifyH.VerifyHistory)

			// 凭证二维码（创意功能）
			authed.GET("/credential/:no/qrcode", qrH.GetQR)

			// 系统日志
			authed.GET("/logs",
				middleware.RoleRequired("school_admin"),
				logH.List)
		}
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"chain":  chainClient != nil && chainClient.IsConnected(),
		})
	})

	// ---------- 5. 启动 ----------
	addr := ":" + cfg.ServerPort
	log.Printf("[HTTP] 服务启动 http://0.0.0.0%s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
