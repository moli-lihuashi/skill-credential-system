package chain

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"skill-credential-backend/config"
)

// CredentialRegistry 合约 ABI（与 contracts/CredentialRegistry.sol 对应）
// 使用手动 ABI 调用，避免依赖 abigen 工具链
const contractABI = `[
	{"inputs":[],"stateMutability":"nonpayable","type":"constructor"},
	{"anonymous":false,"inputs":[{"indexed":true,"internalType":"string","name":"credentialNo","type":"string"},{"indexed":false,"internalType":"string","name":"studentId","type":"string"},{"indexed":false,"internalType":"string","name":"achievementType","type":"string"},{"indexed":false,"internalType":"string","name":"fileHash","type":"string"},{"indexed":true,"internalType":"address","name":"issuer","type":"address"},{"indexed":false,"internalType":"uint256","name":"timestamp","type":"uint256"}],"name":"CredentialStored","type":"event"},
	{"anonymous":false,"inputs":[{"indexed":true,"internalType":"string","name":"credentialNo","type":"string"},{"indexed":false,"internalType":"bool","name":"passed","type":"bool"},{"indexed":false,"internalType":"address","name":"verifier","type":"address"}],"name":"CredentialVerified","type":"event"},
	{"inputs":[{"internalType":"address","name":"nodeAddr","type":"address"}],"name":"addAuthorizedNode","outputs":[],"stateMutability":"nonpayable","type":"function"},
	{"inputs":[{"internalType":"string","name":"studentId","type":"string"}],"name":"registerStudent","outputs":[],"stateMutability":"nonpayable","type":"function"},
	{"inputs":[{"internalType":"string","name":"credentialNo","type":"string"},{"internalType":"string","name":"studentId","type":"string"},{"internalType":"string","name":"achType","type":"string"},{"internalType":"string","name":"title","type":"string"},{"internalType":"string","name":"fileHash","type":"string"},{"internalType":"string","name":"issuerName","type":"string"}],"name":"storeCredential","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"},
	{"inputs":[{"internalType":"string","name":"credentialNo","type":"string"},{"internalType":"string","name":"fileHash","type":"string"}],"name":"verifyCredential","outputs":[{"internalType":"bool","name":"passed","type":"bool"},{"internalType":"string","name":"studentId","type":"string"},{"internalType":"string","name":"achType","type":"string"},{"internalType":"string","name":"title","type":"string"},{"internalType":"uint256","name":"timestamp","type":"uint256"},{"internalType":"string","name":"issuerName","type":"string"}],"stateMutability":"nonpayable","type":"function"},
	{"inputs":[{"internalType":"string","name":"credentialNo","type":"string"}],"name":"getCredential","outputs":[{"internalType":"string","name":"studentId","type":"string"},{"internalType":"string","name":"achType","type":"string"},{"internalType":"string","name":"title","type":"string"},{"internalType":"string","name":"fileHash","type":"string"},{"internalType":"uint256","name":"timestamp","type":"uint256"},{"internalType":"address","name":"issuer","type":"address"},{"internalType":"string","name":"issuerName","type":"string"},{"internalType":"bool","name":"exists","type":"bool"}],"stateMutability":"view","type":"function"},
	{"inputs":[{"internalType":"string","name":"studentId","type":"string"}],"name":"getCredentialNosByStudent","outputs":[{"internalType":"string[]","name":"nos","type":"string[]"}],"stateMutability":"view","type":"function"},
	{"inputs":[],"name":"totalCredentials","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},
	{"inputs":[{"internalType":"string","name":"opHash","type":"string"}],"name":"logOperation","outputs":[],"stateMutability":"nonpayable","type":"function"}
]`

// ChainClient 联盟链客户端
// txMu 串行化「取 nonce → 签名 → 发送」，避免并发请求读到同一 pending nonce
type ChainClient struct {
	client   *ethclient.Client
	contract common.Address
	abi      abi.ABI
	ownerKey *ecdsa.PrivateKey
	chainID  *big.Int
	txMu     sync.Mutex
}

// NewChainClient 创建链客户端
func NewChainClient(cfg *config.Config) (*ChainClient, error) {
	client, err := ethclient.Dial(cfg.ChainRPC)
	if err != nil {
		return nil, fmt.Errorf("连接链节点失败: %w", err)
	}

	parsedABI, err := abi.JSON(strings.NewReader(contractABI))
	if err != nil {
		return nil, fmt.Errorf("解析合约 ABI 失败: %w", err)
	}

	var key *ecdsa.PrivateKey
	if cfg.DeployerKey != "" {
		key, err = crypto.HexToECDSA(strings.TrimPrefix(cfg.DeployerKey, "0x"))
		if err != nil {
			return nil, fmt.Errorf("加载部署者私钥失败: %w", err)
		}
	}

	return &ChainClient{
		client:   client,
		contract: common.HexToAddress(cfg.ContractAddr),
		abi:      parsedABI,
		ownerKey: key,
		chainID:  big.NewInt(cfg.ChainID),
	}, nil
}

// Close 关闭连接
func (c *ChainClient) Close() {
	if c.client != nil {
		c.client.Close()
	}
}

// StoreCredentialResult 上链存证结果
type StoreCredentialResult struct {
	TxHash      string `json:"tx_hash"`
	BlockNumber int64  `json:"block_number"`
	CredentialNo string `json:"credential_no"`
}

// StoreCredential 调用合约上链存证
func (c *ChainClient) StoreCredential(
	credentialNo, studentId, achType, title, fileHash, issuerName string,
) (*StoreCredentialResult, error) {
	if c.ownerKey == nil {
		return nil, fmt.Errorf("未配置链账户私钥")
	}

	// 编码 calldata
	data, err := c.abi.Pack("storeCredential",
		credentialNo, studentId, achType, title, fileHash, issuerName)
	if err != nil {
		return nil, fmt.Errorf("编码合约调用失败: %w", err)
	}

	txHash, blockNo, err := c.sendTx(data)
	if err != nil {
		return nil, err
	}

	return &StoreCredentialResult{
		TxHash:       txHash,
		BlockNumber:  blockNo,
		CredentialNo: credentialNo,
	}, nil
}

// VerifyResult 核验结果
type VerifyResult struct {
	Passed    bool   `json:"passed"`
	StudentId string `json:"student_id"`
	AchType   string `json:"ach_type"`
	Title     string `json:"title"`
	Timestamp uint64 `json:"timestamp"`
	IssuerName string `json:"issuer_name"`
	TxHash    string `json:"tx_hash"`
}

// VerifyCredential 调用合约核验凭证
// verifyCredential 会 emit 事件，需要发交易；交易失败必须显式返回错误，
// 不能吞掉后继续用 eth_call 算出一个看似成功的结果。
func (c *ChainClient) VerifyCredential(credentialNo, fileHash string) (*VerifyResult, error) {
	if c.ownerKey == nil {
		return nil, fmt.Errorf("未配置链账户私钥")
	}

	data, err := c.abi.Pack("verifyCredential", credentialNo, fileHash)
	if err != nil {
		return nil, fmt.Errorf("编码核验调用失败: %w", err)
	}

	txHash, blockNo, err := c.sendTx(data)
	if err != nil {
		return nil, fmt.Errorf("链上核验交易失败: %w", err)
	}
	_ = blockNo

	// 解析返回值：发送交易后通过 Call 重放获取返回
	callMsg := ethereum.CallMsg{
		From: crypto.PubkeyToAddress(c.ownerKey.PublicKey),
		To:   &c.contract,
		Data: data,
	}
	out, err := c.client.CallContract(context.Background(), callMsg, nil)
	if err != nil {
		// 交易型调用可能改状态；改用事件回执解析
		return c.parseVerifyFromReceipt(txHash, credentialNo, fileHash)
	}

	var result struct {
		Passed    bool
		StudentId string
		AchType   string
		Title     string
		Timestamp *big.Int
		IssuerName string
	}
	if err := c.abi.UnpackIntoInterface(&result, "verifyCredential", out); err != nil {
		return c.parseVerifyFromReceipt(txHash, credentialNo, fileHash)
	}

	return &VerifyResult{
		Passed:     result.Passed,
		StudentId:  result.StudentId,
		AchType:    result.AchType,
		Title:      result.Title,
		Timestamp:  result.Timestamp.Uint64(),
		IssuerName: result.IssuerName,
		TxHash:     txHash,
	}, nil
}

// parseVerifyFromReceipt 从交易回执事件解析核验结果（备用路径）
func (c *ChainClient) parseVerifyFromReceipt(txHash, credentialNo, fileHash string) (*VerifyResult, error) {
	// 回退：直接用 getCredential 做本地比对 + 记录交易
	info, err := c.GetCredential(credentialNo)
	if err != nil {
		return nil, err
	}
	// 与合约 _equals（keccak 精确比较）及本地 DB 比较保持一致：大小写敏感
	passed := info.FileHash == fileHash
	return &VerifyResult{
		Passed:     passed,
		StudentId:  info.StudentId,
		AchType:    info.AchType,
		Title:      info.Title,
		Timestamp:  info.Timestamp,
		IssuerName: info.IssuerName,
		TxHash:     txHash,
	}, nil
}

// CredentialInfo 链上存证信息
type CredentialInfo struct {
	StudentId  string `json:"student_id"`
	AchType    string `json:"ach_type"`
	Title      string `json:"title"`
	FileHash   string `json:"file_hash"`
	Timestamp  uint64 `json:"timestamp"`
	Issuer     string `json:"issuer"`
	IssuerName string `json:"issuer_name"`
	Exists     bool   `json:"exists"`
}

// GetCredential 只读查询链上存证
func (c *ChainClient) GetCredential(credentialNo string) (*CredentialInfo, error) {
	data, err := c.abi.Pack("getCredential", credentialNo)
	if err != nil {
		return nil, err
	}

	out, err := c.client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &c.contract,
		Data: data,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("链上查询失败: %w", err)
	}

	var result struct {
		StudentId  string
		AchType    string
		Title      string
		FileHash   string
		Timestamp  *big.Int
		Issuer     common.Address
		IssuerName string
		Exists     bool
	}
	if err := c.abi.UnpackIntoInterface(&result, "getCredential", out); err != nil {
		return nil, fmt.Errorf("解析链上数据失败: %w", err)
	}

	return &CredentialInfo{
		StudentId:  result.StudentId,
		AchType:    result.AchType,
		Title:      result.Title,
		FileHash:   result.FileHash,
		Timestamp:  result.Timestamp.Uint64(),
		Issuer:     result.Issuer.Hex(),
		IssuerName: result.IssuerName,
		Exists:     result.Exists,
	}, nil
}

// LogOperation 操作日志哈希上链
func (c *ChainClient) LogOperation(opHash string) (string, error) {
	if c.ownerKey == nil {
		return "", fmt.Errorf("未配置链账户私钥")
	}
	data, err := c.abi.Pack("logOperation", opHash)
	if err != nil {
		return "", err
	}
	txHash, _, err := c.sendTx(data)
	return txHash, err
}

// sendTx 发送交易并等待回执
func (c *ChainClient) sendTx(data []byte) (string, int64, error) {
	c.txMu.Lock()
	defer c.txMu.Unlock()

	from := crypto.PubkeyToAddress(c.ownerKey.PublicKey)

	nonce, err := c.client.PendingNonceAt(context.Background(), from)
	if err != nil {
		return "", 0, fmt.Errorf("获取 nonce 失败: %w", err)
	}

	gasPrice, err := c.client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", 0, fmt.Errorf("获取 gasPrice 失败: %w", err)
	}

	msg := ethereum.CallMsg{
		From: from,
		To:   &c.contract,
		Data: data,
	}
	gasLimit, err := c.client.EstimateGas(context.Background(), msg)
	if err != nil {
		gasLimit = 1_000_000 // 估算失败时使用默认值
	}

	tx := types.NewTransaction(nonce, c.contract, big.NewInt(0), gasLimit, gasPrice, data)
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(c.chainID), c.ownerKey)
	if err != nil {
		return "", 0, fmt.Errorf("交易签名失败: %w", err)
	}

	if err := c.client.SendTransaction(context.Background(), signedTx); err != nil {
		return "", 0, fmt.Errorf("发送交易失败: %w", err)
	}

	// 等待回执（最多 30 秒）
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	receipt, err := bind.WaitMined(ctx, c.client, signedTx)
	if err != nil {
		return signedTx.Hash().Hex(), 0, fmt.Errorf("等待交易确认超时: %w", err)
	}

	if receipt.Status != types.ReceiptStatusSuccessful {
		return signedTx.Hash().Hex(), receipt.BlockNumber.Int64(),
			fmt.Errorf("交易执行失败(状态=%d)", receipt.Status)
	}

	return signedTx.Hash().Hex(), receipt.BlockNumber.Int64(), nil
}

// IsConnected 检查链连接
func (c *ChainClient) IsConnected() bool {
	if c.client == nil {
		return false
	}
	_, err := c.client.BlockNumber(context.Background())
	return err == nil
}
