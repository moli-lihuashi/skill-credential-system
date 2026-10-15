package handlers

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"skill-credential-backend/config"
	"skill-credential-backend/models"
)

// QRCodeHandler 凭证二维码处理器（创意功能）
// 为每个存证凭证生成可扫描的核验二维码，企业扫码即可跳转核验页
type QRCodeHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewQRCodeHandler(db *gorm.DB, cfg *config.Config) *QRCodeHandler {
	return &QRCodeHandler{db: db, cfg: cfg}
}

// CredentialQR 存证凭证的二维码内容
// 编码为 JSON 格式，前端用 qrcode 库渲染为图片
type CredentialQR struct {
	CredentialNo string `json:"credential_no"`
	StudentName  string `json:"student_name"`
	StudentNo    string `json:"student_no"`
	AchType      string `json:"ach_type"`
	Title        string `json:"title"`
	FileHash     string `json:"file_hash"`
	VerifyURL    string `json:"verify_url"` // 一键跳转核验页
}

// GetQR 返回指定存证凭证的二维码数据（JSON + 人类可读信息）
// 路径: GET /api/credential/:no/qrcode
func (h *QRCodeHandler) GetQR(c *gin.Context) {
	credentialNo := c.Param("no")
	if credentialNo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "缺少存证编号"})
		return
	}

	var cred models.Credential
	if err := h.db.Where("credential_no = ?", credentialNo).First(&cred).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "凭证不存在"})
		return
	}

	// 构造核验直链：扫码后自动跳转到核验页并预填编号
	verifyURL := fmt.Sprintf("/verify?credential_no=%s&file_hash=%s",
		cred.CredentialNo, cred.FileHash)

	qrData := CredentialQR{
		CredentialNo: cred.CredentialNo,
		StudentName:  cred.StudentName,
		StudentNo:    cred.StudentNo,
		AchType:      cred.AchType,
		Title:        cred.Title,
		FileHash:     cred.FileHash,
		VerifyURL:    verifyURL,
	}

	// 将 JSON 编码为 base64，作为二维码内容（更紧凑）
	qrContent := base64.StdEncoding.EncodeToString([]byte(
		fmt.Sprintf("%s|%s|%s|%s", cred.CredentialNo, cred.FileHash, cred.StudentNo, cred.Title),
	))

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"credential": qrData,
			"qr_content": qrContent,
			"verify_url": verifyURL,
		},
	})
}
