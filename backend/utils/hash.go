package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// SHA256String 计算字符串的 SHA-256（返回64位小写十六进制）
func SHA256String(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

// SHA256File 计算文件的 SHA-256
func SHA256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// SHA256Bytes 计算字节切片的 SHA-256
func SHA256Bytes(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// GenerateCredentialNo 生成唯一存证编号
// 格式: CC-YYYYMMDDHHmmss-XXXXXX (CC=Credential of Competence)
func GenerateCredentialNo() string {
	ts := time.Now().Format("20060102150405")
	b := make([]byte, 3)
	rand.Read(b)
	suffix := strings.ToUpper(hex.EncodeToString(b))
	return fmt.Sprintf("CC-%s-%s", ts, suffix)
}

// LogHash 计算日志哈希: SHA256(username|action|detail|timestamp)
func LogHash(username, action, detail string, ts time.Time) string {
	raw := fmt.Sprintf("%s|%s|%s|%d", username, action, detail, ts.Unix())
	return SHA256String(raw)
}
