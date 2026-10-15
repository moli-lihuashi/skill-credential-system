package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"skill-credential-backend/chain"
	"skill-credential-backend/config"
	"skill-credential-backend/models"
	"skill-credential-backend/utils"
)

// VerifyHandler 凭证核验接口
type VerifyHandler struct {
	db    *gorm.DB
	cfg   *config.Config
	chain *chain.ChainClient
}

func NewVerifyHandler(db *gorm.DB, cfg *config.Config, ch *chain.ChainClient) *VerifyHandler {
	return &VerifyHandler{db: db, cfg: cfg, chain: ch}
}

// VerifyRequest 核验请求：存证凭证编号 + 文件哈希（或上传文件）
type VerifyRequest struct {
	CredentialNo string `json:"credential_no" binding:"required"`
	FileHash     string `json:"file_hash" binding:"required"`
	StudentName  string `json:"student_name"` // 可选：学生信息辅助核对
	StudentNo    string `json:"student_no"`
}

// Verify 外部招聘企业核验凭证真实性
// 比对传入文件哈希与链上存储哈希，输出核验结果
func (h *VerifyHandler) Verify(c *gin.Context) {
	var req VerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误: 需要 credential_no + file_hash"})
		return
	}

	employerID := c.GetInt64("user_id") // 可能为 0（游客）
	employerName := c.GetString("username")

	// 1. 先查本地凭证库确认编号存在
	var cred models.Credential
	err := h.db.Where("credential_no = ?", req.CredentialNo).First(&cred).Error
	if err != nil {
		writeLog(h.db, h.chain, employerID, employerName, "VERIFY_FAIL",
			fmt.Sprintf("核验失败 凭证不存在 no=%s", req.CredentialNo))
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "核验不通过（信息存在篡改风险）：凭证编号不存在",
			"data": gin.H{
				"result":        "fail",
				"reason":        "凭证编号不存在",
				"credential_no": req.CredentialNo,
			},
		})
		return
	}

	// 2. 学生信息辅助校验（如果提供了）
	if req.StudentNo != "" && req.StudentNo != cred.StudentNo {
		h.recordVerify(req.CredentialNo, employerID, employerName, req.FileHash, "fail")
		writeLog(h.db, h.chain, employerID, employerName, "VERIFY_FAIL",
			fmt.Sprintf("核验失败 学号不匹配 no=%s", req.CredentialNo))
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"msg":  "核验不通过（信息存在篡改风险）：学号不匹配",
			"data": gin.H{
				"result":        "fail",
				"reason":        "学号不匹配",
				"credential_no": req.CredentialNo,
			},
		})
		return
	}

	// 3. 调用智能合约核验 —— 以链上比对结果为准
	//    防篡改的核心在于链上数据不可改：若此处回退到本地库判定，
	//    则直接改动 MySQL 的 file_hash 就能让核验结果翻转，主张即失效
	chainUnavailable := func(reason string) {
		writeLog(h.db, h.chain, employerID, employerName, "VERIFY_FAIL", reason)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"code": 503,
			"msg":  "链上核验服务暂不可用，请稍后重试",
			"data": gin.H{
				"result":        "error",
				"reason":        "chain_unavailable",
				"credential_no": req.CredentialNo,
			},
		})
	}

	if h.chain == nil {
		chainUnavailable(fmt.Sprintf("链上核验不可用（链客户端未初始化） no=%s", req.CredentialNo))
		return
	}

	chainResult, err := h.chain.VerifyCredential(req.CredentialNo, req.FileHash)
	if err != nil || chainResult == nil {
		chainUnavailable(fmt.Sprintf("链上核验调用失败 no=%s err=%v", req.CredentialNo, err))
		return
	}
	passed := chainResult.Passed

	// 链上结论与本地快照不一致 —— 说明链下记录已被改动，单独留痕告警
	if localMatch := cred.FileHash == req.FileHash; localMatch != passed {
		writeLog(h.db, h.chain, employerID, employerName, "VERIFY_MISMATCH",
			fmt.Sprintf("链上结论与本地记录不一致 no=%s 链上判定=%v 本地比对=%v（链下数据疑被改动）",
				req.CredentialNo, passed, localMatch))
	}

	resultStr := "fail"
	msg := "核验不通过（信息存在篡改风险）"
	if passed {
		resultStr = "pass"
		msg = "核验通过（数据未篡改）"
	}

	// 4. 记录核验日志
	h.recordVerify(req.CredentialNo, employerID, employerName, req.FileHash, resultStr)
	writeLog(h.db, h.chain, employerID, employerName, "VERIFY_"+resultStr,
		fmt.Sprintf("凭证核验 no=%s result=%s", req.CredentialNo, resultStr))

	// 更新核验次数（凭证记录本身不可修改业务字段）
	if passed {
		h.db.Model(&models.Credential{}).Where("id = ?", cred.ID).
			Update("verify_count", gorm.Expr("verify_count + 1"))
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  msg,
		"data": gin.H{
			"result":        resultStr,
			"credential_no": cred.CredentialNo,
			"student_name":  cred.StudentName,
			"student_no":    cred.StudentNo,
			"ach_type":      cred.AchType,
			"title":         cred.Title,
			"file_hash":     cred.FileHash,
			"onchain_time":  cred.OnchainTime,
			"chain_tx_hash": cred.ChainTxHash,
		},
	})
}

// VerifyHistory 核验历史
// 招聘企业只看自己的记录；院校管理员可查全部（审计）；其余角色无权
func (h *VerifyHandler) VerifyHistory(c *gin.Context) {
	employerID := c.GetInt64("user_id")
	role := c.GetString("role")

	var logs []models.VerifyLog
	q := h.db.Order("created_at DESC")
	switch role {
	case models.RoleEmployer:
		q = q.Where("employer_id = ?", employerID)
	case models.RoleSchoolAdmin:
		// 管理端审计
	default:
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "无权查看核验历史"})
		return
	}
	if err := q.Limit(100).Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": logs})
}

// recordVerify 保存核验记录
func (h *VerifyHandler) recordVerify(no string, empID int64, empName, hash, result string) {
	h.db.Create(&models.VerifyLog{
		CredentialNo: no,
		EmployerID:   empID,
		EmployerName: empName,
		InputHash:    hash,
		ChainResult:  result,
		CreatedAt:    time.Now(),
	})
}

// ========== 成果录入 ==========

// AchievementHandler 成果管理
type AchievementHandler struct {
	db    *gorm.DB
	cfg   *config.Config
	chain *chain.ChainClient
}

func NewAchievementHandler(db *gorm.DB, cfg *config.Config, ch *chain.ChainClient) *AchievementHandler {
	return &AchievementHandler{db: db, cfg: cfg, chain: ch}
}

// CreateAchievementRequest 录入成果请求
type CreateAchievementRequest struct {
	StudentID   int64  `json:"student_id" binding:"required"`
	Title       string `json:"title" binding:"required"`
	AchType     string `json:"ach_type" binding:"required,oneof=certificate competition internship project"`
	Description string `json:"description"`
	Issuer      string `json:"issuer"`
	FileHash    string `json:"file_hash" binding:"required"`
	FilePath    string `json:"file_path"`
}

// Create 录入成果（院校管理员：证书/竞赛；实训企业：实训评价/项目）
func (h *AchievementHandler) Create(c *gin.Context) {
	var req CreateAchievementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误: " + err.Error()})
		return
	}

	role := c.GetString("role")
	if role != models.RoleSchoolAdmin && role != models.RoleCompany {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "只有院校管理员和实训企业可录入成果"})
		return
	}

	// 角色-类型权限：院校→证书/竞赛；企业→实训/项目
	if role == models.RoleSchoolAdmin &&
		req.AchType != models.AchCertificate && req.AchType != models.AchCompetition {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "院校管理员只能录入技能证书、竞赛获奖"})
		return
	}
	if role == models.RoleCompany &&
		req.AchType != models.AchInternship && req.AchType != models.AchProject {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "实训企业只能录入岗位实训评价、实训项目成果"})
		return
	}

	// 目标学生必须存在且为 student 角色，防止挂到任意/不存在的用户
	var student models.User
	if err := h.db.First(&student, req.StudentID).Error; err != nil ||
		student.Role != models.RoleStudent {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "目标学生不存在或角色不正确"})
		return
	}

	ach := models.Achievement{
		StudentID:   req.StudentID,
		Title:       req.Title,
		AchType:     req.AchType,
		Description: req.Description,
		Issuer:      req.Issuer,
		FileHash:    req.FileHash,
		FilePath:    req.FilePath,
		Status:      "draft",
		CreatedBy:   c.GetInt64("user_id"),
	}

	if err := h.db.Create(&ach).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "创建成果失败"})
		return
	}

	writeLog(h.db, h.chain, c.GetInt64("user_id"), c.GetString("username"), "CREATE_ACHIEVEMENT",
		fmt.Sprintf("录入成果 title=%s type=%s studentId=%d", req.Title, req.AchType, req.StudentID))

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "录入成功", "data": ach})
}

// ListAchievements 成果列表
// 可见范围按角色收敛：学生看本人；院校/实训企业看全部；招聘企业无权查看
func (h *AchievementHandler) List(c *gin.Context) {
	role := c.GetString("role")
	userID := c.GetInt64("user_id")

	q := h.db.Order("created_at DESC")

	switch role {
	case models.RoleStudent:
		q = q.Where("student_id = ?", userID)
	case models.RoleSchoolAdmin:
		// 管理端可见全部
	case models.RoleCompany:
		// 实训企业只管自己录入的实训/项目成果，避免看到院校证书/竞赛与其他企业记录
		q = q.Where("created_by = ? AND ach_type IN ?", userID,
			[]string{models.AchInternship, models.AchProject})
	default:
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "无权查看成果列表"})
		return
	}

	// 可选按类型筛选
	if t := c.Query("ach_type"); t != "" {
		q = q.Where("ach_type = ?", t)
	}

	var list []models.Achievement
	if err := q.Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": list})
}

// ========== 系统日志 ==========

// LogHandler 日志接口
type LogHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewLogHandler(db *gorm.DB, cfg *config.Config) *LogHandler {
	return &LogHandler{db: db, cfg: cfg}
}

// List 系统日志（院校管理员可查）
func (h *LogHandler) List(c *gin.Context) {
	var logs []models.SysLog
	if err := h.db.Order("created_at DESC").Limit(200).Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": logs})
}

// ========== 文件上传 ==========

// UploadHandler 文件上传（返回文件哈希）
type UploadHandler struct {
	db    *gorm.DB
	cfg   *config.Config
	chain *chain.ChainClient
}

func NewUploadHandler(db *gorm.DB, cfg *config.Config, ch *chain.ChainClient) *UploadHandler {
	return &UploadHandler{db: db, cfg: cfg, chain: ch}
}

// Upload 上传附件文件，系统计算文件哈希
func (h *UploadHandler) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请选择文件"})
		return
	}

	if file.Size > 20*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "文件不能超过 20MB"})
		return
	}

	// 保存文件
	savePath := fmt.Sprintf("%s/%d_%s", h.cfg.UploadDir, time.Now().UnixNano(), file.Filename)
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "保存文件失败"})
		return
	}

	// 计算 SHA-256 哈希
	fileHash, err := utils.SHA256File(savePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "计算文件哈希失败"})
		return
	}

	writeLog(h.db, h.chain, c.GetInt64("user_id"), c.GetString("username"), "UPLOAD_FILE",
		fmt.Sprintf("上传附件 name=%s hash=%s", file.Filename, fileHash))

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "上传成功",
		"data": gin.H{
			"file_name": file.Filename,
			"file_hash": fileHash,
			"file_path": savePath,
			"file_size": file.Size,
		},
	})
}
