package models

import "time"

// ========== 用户 ==========
const (
	RoleStudent    = "student"     // 学生
	RoleSchoolAdmin = "school_admin" // 院校管理员
	RoleCompany    = "company"     // 实训企业
	RoleEmployer   = "employer"    // 招聘企业
)

type User struct {
	ID            int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Username      string    `gorm:"size:64;uniqueIndex" json:"username"`
	Password      string    `gorm:"size:255" json:"-"`
	RealName      string    `gorm:"size:64" json:"real_name"`
	// StudentNo 对学生角色唯一，由 Register 在应用层校验
	// （不能建全局唯一索引：非学生角色 student_no 为空串会互相冲突）
	StudentNo     string    `gorm:"size:32;index" json:"student_no"`
	Role          string    `gorm:"size:32;index" json:"role"`
	OrgName       string    `gorm:"size:128" json:"org_name"`
	EthAddress    string    `gorm:"size:42" json:"eth_address"`
	EthPrivateKey string    `gorm:"size:256" json:"-"`
	Status        int8      `gorm:"default:1" json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

// ========== 成果 ==========
const (
	AchCertificate  = "certificate"  // 技能证书
	AchCompetition  = "competition"  // 竞赛获奖
	AchInternship   = "internship"   // 岗位实训
	AchProject      = "project"      // 项目实践
)

type Achievement struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	StudentID   int64     `gorm:"index" json:"student_id"`
	Title       string    `gorm:"size:128" json:"title"`
	AchType     string    `gorm:"size:32;index" json:"ach_type"`
	Description string    `gorm:"type:text" json:"description"`
	Issuer      string    `gorm:"size:128" json:"issuer"`
	FileHash    string    `gorm:"size:64" json:"file_hash"`
	FilePath    string    `gorm:"size:255" json:"file_path"`
	Status      string    `gorm:"size:16;index" json:"status"` // draft/submitted/onchain
	CreatedBy   int64     `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
}

// ========== 存证凭证（只增不改删） ==========
type Credential struct {
	ID             int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	CredentialNo   string    `gorm:"size:64;uniqueIndex" json:"credential_no"`
	AchievementID  int64     `json:"achievement_id"`
	StudentID      int64     `json:"student_id"`
	StudentName    string    `gorm:"size:64" json:"student_name"`
	StudentNo      string    `gorm:"size:32;index" json:"student_no"`
	AchType        string    `gorm:"size:32;index" json:"ach_type"`
	Title          string    `gorm:"size:128" json:"title"`
	ChainTxHash    string    `gorm:"size:66" json:"chain_tx_hash"`
	ChainBlockNo   int64     `json:"chain_block_no"`
	FileHash       string    `gorm:"size:64" json:"file_hash"`
	IssuerAddr     string    `gorm:"size:42" json:"issuer_addr"`
	OnchainTime    time.Time `gorm:"index" json:"onchain_time"`
	VerifyCount    int       `gorm:"default:0" json:"verify_count"`
	Status         int8      `gorm:"default:1" json:"status"`
}

// ========== 核验记录 ==========
type VerifyLog struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	CredentialNo string    `gorm:"size:64;index" json:"credential_no"`
	EmployerID   int64     `json:"employer_id"`
	EmployerName string    `gorm:"size:128" json:"employer_name"`
	InputHash    string    `gorm:"size:64" json:"input_hash"`
	ChainResult  string    `gorm:"size:16" json:"chain_result"` // pass/fail
	CreatedAt    time.Time `json:"created_at"`
}

// ========== 系统日志 ==========
type SysLog struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     int64     `json:"user_id"`
	Username   string    `gorm:"size:64" json:"username"`
	Action     string    `gorm:"size:64;index" json:"action"`
	Detail     string    `gorm:"type:text" json:"detail"`
	LogHash    string    `gorm:"size:64" json:"log_hash"`
	BatchHash  string    `gorm:"size:80" json:"batch_hash"`
	Onchain    int8      `gorm:"default:0;index" json:"onchain"`
	CreatedAt  time.Time `json:"created_at"`
}
