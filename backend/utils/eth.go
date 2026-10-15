package utils

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
)

// EthAccount 以太坊账户
// PrivateKey 仅供服务端内部签名使用，禁止进入 JSON 响应（无 json 标签）
type EthAccount struct {
	Address    string `json:"address"`
	PrivateKey string // 十六进制，不含 0x 前缀；不序列化
}

// GenerateEthAccount 生成以太坊账户（公私钥对）
// 注册时自动为每个用户生成链账户
func GenerateEthAccount() (*EthAccount, error) {
	key, err := crypto.GenerateKey()
	if err != nil {
		return nil, fmt.Errorf("生成链账户失败: %w", err)
	}
	addr := crypto.PubkeyToAddress(key.PublicKey)
	return &EthAccount{
		Address:    addr.Hex(),
		PrivateKey: fmt.Sprintf("%x", crypto.FromECDSA(key)),
	}, nil
}

// PrivateKeyToAddress 由私钥推导地址
func PrivateKeyToAddress(privHex string) (string, error) {
	key, err := crypto.HexToECDSA(privHex)
	if err != nil {
		return "", err
	}
	return crypto.PubkeyToAddress(key.PublicKey).Hex(), nil
}

// LoadPrivateKey 加载私钥
func LoadPrivateKey(privHex string) (*ecdsa.PrivateKey, error) {
	return crypto.HexToECDSA(privHex)
}
