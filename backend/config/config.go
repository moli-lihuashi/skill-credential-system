package config

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"
)

// Config 系统配置：开发环境默认值，可通过环境变量覆盖
type Config struct {
	// MySQL
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// 链节点（开发：Ganache http://127.0.0.1:8545）
	ChainRPC     string
	ChainID      int64
	ContractAddr string // 部署后的合约地址
	// 合约部署者私钥（演示环境，生产应使用密钥管理服务）
	DeployerKey string

	// JWT
	JWTSecret string

	// 服务端口
	ServerPort string

	// 附件上传目录
	UploadDir string
}

// Load 加载配置
// 敏感项（JWT_SECRET / DEPLOYER_KEY）不再使用源码内置默认值：
// 未显式配置时 fail-fast，避免线上落入公开已知密钥。
func Load() *Config {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("[CONFIG] 未设置 JWT_SECRET：请通过环境变量提供强随机密钥，禁止使用内置默认值")
	}
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		log.Fatal("[CONFIG] 未设置 DB_PASSWORD：请通过环境变量提供数据库口令")
	}
	deployerKey := os.Getenv("DEPLOYER_KEY")
	if deployerKey == "" {
		// 链客户端在未配置合约时可不启动；此处仅生成占位，避免空串进入签名流程
		buf := make([]byte, 32)
		if _, err := rand.Read(buf); err != nil {
			log.Fatalf("[CONFIG] 生成占位密钥失败: %v", err)
		}
		log.Println("[CONFIG] 警告: 未设置 DEPLOYER_KEY，上链功能不可用（已生成随机占位）")
		deployerKey = hex.EncodeToString(buf)
	}

	return &Config{
		DBHost:       getEnv("DB_HOST", "127.0.0.1"),
		DBPort:       getEnv("DB_PORT", "3306"),
		DBUser:       getEnv("DB_USER", "root"),
		DBPassword:   dbPassword,
		DBName:       getEnv("DB_NAME", "skill_credential"),
		ChainRPC:     getEnv("CHAIN_RPC", "http://127.0.0.1:8545"),
		ChainID:      1337,
		ContractAddr: getEnv("CONTRACT_ADDR", ""),
		DeployerKey:  deployerKey,
		JWTSecret:    jwtSecret,
		ServerPort:   getEnv("SERVER_PORT", "8080"),
		UploadDir:    getEnv("UPLOAD_DIR", "./uploads"),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

