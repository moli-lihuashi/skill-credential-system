package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"skill-credential-backend/chain"
	"skill-credential-backend/config"
	"skill-credential-backend/middleware"
	"skill-credential-backend/models"
	"skill-credential-backend/utils"
)

type AuthHandler struct {
	db    *gorm.DB
	cfg   *config.Config
	chain *chain.ChainClient
}

func NewAuthHandler(db *gorm.DB, cfg *config.Config, ch *chain.ChainClient) *AuthHandler {
	return &AuthHandler{db: db, cfg: cfg, chain: ch}
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username  string `json:"username" binding:"required,min=3,max=64"`
	Password  string `json:"password" binding:"required,min=6,max=64"`
	RealName  string `json:"real_name" binding:"required"`
	Role      string `json:"role" binding:"required,oneof=student school_admin company employer"`
	StudentNo string `json:"student_no"`
	OrgName   string `json:"org_name"`
}

// Register 注册：自动生成链账户地址
// 安全策略：
//  1. 公开注册仅 student/employer 立即生效；school_admin/company 需管理员审核（Status=0）
//  2. 学号在学生角色内唯一（LookupStudent/Store 依赖它定位学生）
//  3. 不持久化链私钥——上链签名使用服务端 DEPLOYER_KEY，用户私钥入库只会扩大泄露面
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误: " + err.Error()})
		return
	}

	// 学生角色必须填学号
	if req.Role == models.RoleStudent && req.StudentNo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "学生角色必须填写学号"})
		return
	}

	// 学号唯一：重复学号会导致成果/凭证挂错学生
	if req.Role == models.RoleStudent {
		var cnt int64
		if err := h.db.Model(&models.User{}).
			Where("student_no = ? AND role = ?", req.StudentNo, models.RoleStudent).
			Count(&cnt).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询失败"})
			return
		}
		if cnt > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "该学号已被注册"})
			return
		}
	}

	// 检查用户名唯一
	var count int64
	h.db.Model(&models.User{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "用户名已存在"})
		return
	}

	// 密码加密
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "密码加密失败"})
		return
	}

	// 自动生成区块链账户地址（仅落库地址，私钥不持久化）
	ethAcc, err := utils.GenerateEthAccount()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "生成链账户失败: " + err.Error()})
		return
	}

	// 高权限角色需审核后激活，防止任意人自选 school_admin/company 伪造上链
	status := int8(1)
	pendingMsg := "注册成功"
	if req.Role == models.RoleSchoolAdmin || req.Role == models.RoleCompany {
		status = 0
		pendingMsg = "注册成功，请联系管理员审核激活后再登录"
	}

	user := models.User{
		Username:   req.Username,
		Password:   string(hashed),
		RealName:   req.RealName,
		StudentNo:  req.StudentNo,
		Role:       req.Role,
		OrgName:    req.OrgName,
		EthAddress: ethAcc.Address,
		Status:     status,
	}

	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "创建用户失败: " + err.Error()})
		return
	}

	// 记录日志
	writeLog(h.db, h.chain, user.ID, user.Username, "REGISTER",
		"用户注册 role="+req.Role+" addr="+ethAcc.Address)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  pendingMsg,
		"data": gin.H{
			"user_id":     user.ID,
			"username":    user.Username,
			"role":        user.Role,
			"eth_address": user.EthAddress,
			"pending":     status == 0,
		},
	})
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login 登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	var user models.User
	if err := h.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "用户名或密码错误"})
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "用户名或密码错误"})
		return
	}

	if user.Status != 1 {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "账号已被禁用"})
		return
	}

	token, err := middleware.GenerateToken(h.cfg.JWTSecret, user.ID, user.Username, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "生成令牌失败"})
		return
	}

	writeLog(h.db, h.chain, user.ID, user.Username, "LOGIN", "用户登录")

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "登录成功",
		"data": gin.H{
			"token":       token,
			"user_id":     user.ID,
			"username":    user.Username,
			"real_name":   user.RealName,
			"role":        user.Role,
			"org_name":    user.OrgName,
			"student_no":  user.StudentNo,
			"eth_address": user.EthAddress,
		},
	})
}

// Profile 获取个人信息（含链账户地址）
func (h *AuthHandler) Profile(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"user_id":     user.ID,
			"username":    user.Username,
			"real_name":   user.RealName,
			"role":        user.Role,
			"student_no":  user.StudentNo,
			"org_name":    user.OrgName,
			"eth_address": user.EthAddress,
			"created_at":  user.CreatedAt,
		},
	})
}

// LookupStudent 按学号查询学生（成果录入表单自动关联，免手工填学生 ID）
// GET /api/user/lookup-student?student_no=xxx
func (h *AuthHandler) LookupStudent(c *gin.Context) {
	studentNo := c.Query("student_no")
	if studentNo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请提供学号"})
		return
	}

	role := c.GetString("role")
	if role != models.RoleSchoolAdmin && role != models.RoleCompany {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "仅院校管理员和实训企业可查询学生"})
		return
	}

	var user models.User
	if err := h.db.Where("student_no = ? AND role = ?", studentNo, models.RoleStudent).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "未找到该学号对应的学生"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"user_id":    user.ID,
			"real_name":  user.RealName,
			"student_no": user.StudentNo,
		},
	})
}
