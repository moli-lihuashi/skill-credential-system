package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"skill-credential-backend/chain"
	"skill-credential-backend/config"
	"skill-credential-backend/models"
	"skill-credential-backend/utils"
)

// writeLog 写操作日志（含日志哈希），不阻塞请求
// 日志哈希上链由后台 worker 定期批量提交（对应论文 5.3.7「定期批量上链」）
func writeLog(db *gorm.DB, chainClient *chain.ChainClient, userID int64, username, action, detail string) {
	ts := time.Now()
	logHash := utils.LogHash(username, action, detail, ts)
	db.Create(&models.SysLog{
		UserID:   userID,
		Username: username,
		Action:   action,
		Detail:   detail,
		LogHash:  logHash,
		Onchain:  0,
	})
	// 不再同步调用链上 logOperation，由 StartLogOnChainWorker 异步批量处理
}

// StartLogOnChainWorker 后台协程：定期将未上链的日志哈希批量写入区块链
// 对应论文 5.3.7「日志哈希定期批量上链调用合约的 logOperation 函数」
func StartLogOnChainWorker(db *gorm.DB, ch *chain.ChainClient, intervalSec int) {
	if ch == nil {
		log.Println("[LOG-WORKER] 链客户端未初始化，日志上链功能不可用")
		return
	}
	if intervalSec <= 0 {
		intervalSec = 30 // 默认 30 秒批量一次
	}
	log.Printf("[LOG-WORKER] 启动日志批量上链协程，间隔 %d 秒", intervalSec)

	go func() {
		ticker := time.NewTicker(time.Duration(intervalSec) * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			// 查询未上链的日志（每次最多 50 条，避免单批过大）
			var pending []models.SysLog
			if err := db.Where("onchain = ?", 0).Order("id asc").Limit(50).Find(&pending).Error; err != nil {
				log.Printf("[LOG-WORKER] 查询待上链日志失败: %v", err)
				continue
			}
			if len(pending) == 0 {
				continue
			}

			success := 0
			fail := 0
			for _, rec := range pending {
				txHash, err := ch.LogOperation(rec.LogHash)
				if err != nil {
					fail++
					log.Printf("[LOG-WORKER] 日志上链失败 id=%d: %v", rec.ID, err)
					continue // 失败的留下一批重试
				}
				if err := db.Model(&rec).Updates(map[string]interface{}{
					"batch_hash": txHash,
					"onchain":    1,
				}).Error; err != nil {
					log.Printf("[LOG-WORKER] 更新上链状态失败 id=%d: %v", rec.ID, err)
				}
				success++
			}
			log.Printf("[LOG-WORKER] 批量上链完成: 成功=%d 失败=%d", success, fail)
		}
	}()
}

// CredentialHandler 存证相关接口
type CredentialHandler struct {
	db    *gorm.DB
	cfg   *config.Config
	chain *chain.ChainClient
}

func NewCredentialHandler(db *gorm.DB, cfg *config.Config, ch *chain.ChainClient) *CredentialHandler {
	return &CredentialHandler{db: db, cfg: cfg, chain: ch}
}

// StoreRequest 上链存证请求
type StoreRequest struct {
	AchievementID int64  `json:"achievement_id" binding:"required"`
}

// Store 上链存证：调用智能合约，返回唯一存证凭证编号
func (h *CredentialHandler) Store(c *gin.Context) {
	var req StoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	// 链客户端未初始化时（CONTRACT_ADDR 为空或链连接失败）明确 503，避免空指针 panic
	if h.chain == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "msg": "链服务未就绪，请稍后重试或联系管理员"})
		return
	}

	userID := c.GetInt64("user_id")
	username := c.GetString("username")
	role := c.GetString("role")

	// 只有院校管理员和实训企业有上链权限
	if role != models.RoleSchoolAdmin && role != models.RoleCompany {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "只有院校管理员和实训企业拥有成果上链权限"})
		return
	}

	// 查成果
	var ach models.Achievement
	if err := h.db.First(&ach, req.AchievementID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "成果不存在"})
		return
	}

	if ach.Status == "onchain" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "该成果已上链存证"})
		return
	}

	// 查学生
	var student models.User
	if err := h.db.First(&student, ach.StudentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "学生不存在"})
		return
	}

	// 文件哈希校验
	fileHash := ach.FileHash
	if fileHash == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "附件文件哈希为空，请先上传附件"})
		return
	}

	// 生成唯一存证编号
	credentialNo := utils.GenerateCredentialNo()

	// 调用智能合约上链
	issuerName := username
	if u, err := getUserName(h.db, userID); err == nil {
		issuerName = u
	}

	result, err := h.chain.StoreCredential(
		credentialNo,
		student.StudentNo,
		ach.AchType,
		ach.Title,
		fileHash,
		issuerName,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "上链失败: " + err.Error(),
		})
		return
	}

	// 写入存证凭证表（只增不改删）
	cred := models.Credential{
		CredentialNo:  credentialNo,
		AchievementID: ach.ID,
		StudentID:     student.ID,
		StudentName:   student.RealName,
		StudentNo:     student.StudentNo,
		AchType:       ach.AchType,
		Title:         ach.Title,
		ChainTxHash:   result.TxHash,
		ChainBlockNo:  result.BlockNumber,
		FileHash:      fileHash,
		IssuerAddr:    "",
		OnchainTime:   time.Now(),
		Status:        1,
	}
	if err := h.db.Create(&cred).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "保存凭证记录失败"})
		return
	}

	// 更新成果状态
	h.db.Model(&ach).Update("status", "onchain")

	// 操作日志 + 日志哈希上链
	detail := fmt.Sprintf("成果上链 credentialNo=%s achId=%d fileHash=%s",
		credentialNo, ach.ID, fileHash)
	writeLog(h.db, h.chain, userID, username, "STORE_CREDENTIAL", detail)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "上链存证成功",
		"data": gin.H{
			"credential_no":  credentialNo,
			"chain_tx_hash":  result.TxHash,
			"chain_block_no": result.BlockNumber,
			"file_hash":      fileHash,
			"student_name":   student.RealName,
			"student_no":     student.StudentNo,
			"ach_type":       ach.AchType,
			"title":          ach.Title,
		},
	})
}

// MyCredentials 查看存证历史
// 学生：本人全部；院校管理员：全部（审计）；实训企业：本人录入成果的存证
func (h *CredentialHandler) MyCredentials(c *gin.Context) {
	userID := c.GetInt64("user_id")
	role := c.GetString("role")

	var creds []models.Credential
	q := h.db.Order("onchain_time DESC")

	switch role {
	case models.RoleStudent:
		q = q.Where("student_id = ?", userID)
	case models.RoleSchoolAdmin:
		// 管理端可看全部
	case models.RoleCompany:
		// 实训企业只看自己录入的成果所生成的存证
		q = q.Where("achievement_id IN (?)",
			h.db.Model(&models.Achievement{}).Select("id").Where("created_by = ?", userID))
	default:
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "无权查看"})
		return
	}

	// 可选分页：page/page_size 同时提供时返回分页结构，否则保持兼容返回全量
	page, pageSize := parsePage(c)
	if page > 0 && pageSize > 0 {
		var total int64
		if err := q.Count(&total).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询失败"})
			return
		}
		if err := q.Offset((page - 1) * pageSize).Limit(pageSize).Find(&creds).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 200, "data": creds, "total": total, "page": page, "page_size": pageSize})
		return
	}

	if err := q.Find(&creds).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": creds})
}

// SearchRequest 检索请求
type SearchRequest struct {
	StudentName string `form:"student_name"`
	StudentNo   string `form:"student_no"`
	AchType     string `form:"ach_type"`
	Page        int    `form:"page"`
	PageSize    int    `form:"page_size"`
}

// Search 按学生姓名/学号/成果类型检索存证凭证（服务端分页）
func (h *CredentialHandler) Search(c *gin.Context) {
	var req SearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	q := h.db.Model(&models.Credential{})
	if req.StudentName != "" {
		q = q.Where("student_name LIKE ?", "%"+req.StudentName+"%")
	}
	if req.StudentNo != "" {
		q = q.Where("student_no = ?", req.StudentNo)
	}
	if req.AchType != "" {
		q = q.Where("ach_type = ?", req.AchType)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "检索失败"})
		return
	}

	page, pageSize := req.Page, req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	var creds []models.Credential
	if err := q.Order("onchain_time DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&creds).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "检索失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200, "data": creds,
		"total": total, "page": page, "page_size": pageSize,
	})
}

// parsePage 解析可选分页参数；未提供时返回 0,0 表示不分页
func parsePage(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	if page < 0 {
		page = 0
	}
	if pageSize < 0 {
		pageSize = 0
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func getUserName(db *gorm.DB, userID int64) (string, error) {
	var u models.User
	err := db.Select("real_name").First(&u, userID).Error
	return u.RealName, err
}
