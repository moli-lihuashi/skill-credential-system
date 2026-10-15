# 基于联盟链的高校毕业生技能成果可信存证系统

> 毕业设计作品

## 项目简介

针对高校学生实训报告、技能证书、竞赛获奖等成果容易伪造、纸质证书核验繁琐的痛点，搭建面向职业院校的**联盟链可信存证平台**。学校与实训企业作为联盟链授权节点，将成果哈希上链存证；招聘企业可通过存证编号进行真实性核验。

## 技术栈

| 层次   | 技术                                       |
| ---- | ---------------------------------------- |
| 智能合约 | Solidity 0.8.19                          |
| 后端服务 | Golang (Gin + GORM + go-ethereum)        |
| 前端框架 | Vue3 + Element Plus + Vue Router + Pinia |
| 数据库  | MySQL 5.7+                               |
| 区块链  | Ganache 本地联盟链（可扩展私链）                     |
| 部署环境 | Linux CentOS7 + Nginx                    |

## 系统功能

1. **多角色注册与链身份管理** — 支持学生、院校管理员、实训企业、招聘企业 4 类角色；注册自动分配链账户
2. **技能成果上链存证** — 院校/企业录入成果（学号自动关联学生），文件哈希写入智能合约，返回唯一存证编号
3. **成果凭证核验查询** — 外部企业凭编号+哈希核验，输出「核验通过/不通过」；支持上传文件自动算哈希、扫码/链接预填、一键篡改演示
4. **存证数据管理** — 链上记录不可删改；按姓名/学号/类型检索；操作日志哈希上链

### 体验与创意亮点

- 学生端生成**可扫码核验二维码**，扫码自动预填存证参数；一键复制核验链接分享给招聘企业
- 公开核验页浏览器端计算 SHA-256，支持「填入演示数据 / 一键篡改演示 / 打印导出」
- 登录页演示账号一键填充，答辩演示无需手输
- 全站统一品牌视觉（藏蓝 + 语义色设计令牌），表格分页与复制交互

## 目录结构

```
skill-credential-system/
├── contracts/           # Solidity 智能合约
│   ├── CredentialRegistry.sol
│   └── deploy.js        # 合约部署脚本
├── backend/             # Golang 后端服务
│   ├── main.go          # 入口
│   ├── config/          # 配置
│   ├── models/          # 数据模型
│   ├── handlers/        # API 处理器
│   ├── middleware/       # JWT 鉴权
│   ├── chain/           # 联盟链客户端
│   └── utils/           # 工具（哈希/链账户）
├── frontend/            # Vue3 前端
│   └── src/
│       ├── views/       # 四端页面
│       ├── layouts/     # 布局
│       ├── router/      # 路由权限
│       └── api/         # HTTP 客户端
├── sql/init.sql         # MySQL 初始化脚本
├── deploy/              # 部署与测试脚本
└── docs/                # 系统设计文档
```

## 快速开始（开发环境）

> **安全提示**：敏感项必须通过环境变量注入，源码中不再提供默认值：  
> `JWT_SECRET`（JWT 签名）、`DB_PASSWORD`（数据库口令）、`DEPLOYER_KEY`（部署者私钥）。  
> 一键脚本 `deploy/start.sh` 优先读环境变量，其次读本地 `deploy/.env.demo`（可复制 `deploy/.env.demo.example` 生成，已被 .gitignore 忽略）。  
> 演示账号口令同样不写入源码：登录页「一键填充」的口令由构建时变量 `VITE_DEMO_PASSWORD` 注入（见 `frontend/.env`，未提交），未配置时只填充用户名。  
> 公开注册仅 `student`/`employer` 立即生效；`school_admin`/`company` 注册后需管理员审核激活。

### 1. 启动本地链

```bash
npm install -g ganache
mkdir -p ./chain-data
ganache --port 8545 --chainId 1337 --db ./chain-data
```

### 2. 部署合约

```bash
cd contracts
npm install ethers solc
node deploy.js
# 记录输出的合约地址
```

### 3. 初始化数据库

```bash
mysql -uroot -p < sql/init.sql
```

### 4. 启动后端

```bash
cd backend
go mod tidy
export CONTRACT_ADDR=<上一步的合约地址>
export CHAIN_RPC=http://127.0.0.1:8545
go run main.go
```

### 5. 启动前端

```bash
cd frontend
npm install
npm run dev
# 访问 http://localhost:5173
```

## CentOS7 部署

```bash
bash deploy/deploy_centos7.sh
```

## 测试

```bash
bash deploy/test_all.sh
```

测试覆盖：角色注册、登录鉴权、成果录入、上链存证、篡改模拟核验、并发核验。

## 四类角色

| 角色    | 功能               | 上链权限 |
| ----- | ---------------- | ---- |
| 学生    | 查看个人存证历史         | 无    |
| 院校管理员 | 录入证书/竞赛、上链、检索、日志 | 有    |
| 实训企业  | 录入实训评价/项目、上链     | 有    |
| 招聘企业  | 凭证核验             | 无    |
