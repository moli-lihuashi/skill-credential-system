#!/bin/bash
# ============================================================
# 重建区块链环境 + 重新部署合约 + 重启服务
# 用法: bash /opt/skill-credential/deploy/rebuild_chain.sh
#
# 依赖环境变量：DB_PASSWORD / MYSQL_ROOT_PW / DEPLOYER_KEY / JWT_SECRET
# ============================================================
set -u
set -o pipefail
export PATH=/usr/local/go/bin:/usr/local/nodejs/bin:$PATH
export LANG=zh_CN.UTF-8
export LC_ALL=zh_CN.UTF-8

: "${DB_PASSWORD:?请通过环境变量提供 DB_PASSWORD}"
: "${DEPLOYER_KEY:?请通过环境变量提供 DEPLOYER_KEY}"
: "${JWT_SECRET:?请通过环境变量提供 JWT_SECRET}"
MYSQL_CMD_PW="${MYSQL_ROOT_PW:-$DB_PASSWORD}"

echo "================================================"
echo "  区块链环境重建脚本"
echo "================================================"

# ---------- ② 停服 + 备份 ----------
echo ""
echo "[1/6] 停服 + 数据库备份..."
systemctl stop nginx 2>/dev/null
pkill -f skill-credential-server 2>/dev/null
pkill -f ganache 2>/dev/null
sleep 2
BACKUP_FILE="/root/db_backup_$(date +%m%d).sql"
mysqldump -uroot -p"${MYSQL_CMD_PW}" skill_credential > "$BACKUP_FILE" 2>/dev/null
echo "  数据库已备份到: $BACKUP_FILE"

# ---------- ③ 清空链目录 ----------
echo ""
echo "[2/6] 清空链数据目录..."
rm -rf /opt/skill-credential/chain-data
mkdir -p /opt/skill-credential/chain-data
echo "  链目录已清空"

# ---------- ④ 清链相关表 ----------
echo ""
echo "[3/6] 清理链相关数据表..."
mysql -uroot -p"${MYSQL_CMD_PW}" skill_credential << 'SQL'
TRUNCATE TABLE credentials;
TRUNCATE TABLE verify_logs;
TRUNCATE TABLE sys_logs;
UPDATE achievements SET status='draft' WHERE status='onchain';
SQL
echo "  credentials / verify_logs / sys_logs 已清空"
echo "  achievements 已重置为 draft"

# ---------- ⑤ 起链 + 部署合约 ----------
echo ""
echo "[4/6] 启动 Ganache 链..."
nohup ganache --host 0.0.0.0 --port 8545 -i 1337 -a 10 -e 1000 \
    --wallet.deterministic --db /opt/skill-credential/chain-data \
    > /tmp/ganache.log 2>&1 &
sleep 6
if curl -s -X POST -H 'Content-Type: application/json' \
    --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
    http://127.0.0.1:8545 | grep -q "result"; then
    echo "  Ganache 已启动 (带持久化 --db)"
else
    echo "  Ganache 启动失败！"
    exit 1
fi

echo ""
echo "[5/6] 部署智能合约..."
cd /opt/skill-credential/contracts || exit 1
# 部署者私钥仅从环境变量注入，禁止硬编码
export DEPLOYER_KEY
node deploy.js 2>&1 | tee /tmp/deploy_output.txt

# 从部署输出提取合约地址
NEW_ADDR=$(grep -o 'CONTRACT_ADDR=0x[a-fA-F0-9]\{40\}' /tmp/deploy_output.txt | head -1 | cut -d= -f2)
if [ -z "$NEW_ADDR" ]; then
    NEW_ADDR=$(grep -o '0x[a-fA-F0-9]\{40\}' /tmp/deploy_output.txt | tail -1)
fi

EXPECTED="0xe78A0F7E598Cc8b0Bb87894B0F60dD2a88d6a8Ab"
echo ""
if [ "$NEW_ADDR" = "$EXPECTED" ]; then
    echo "  合约地址确认: $NEW_ADDR (与预期一致)"
else
    echo "  ⚠️ 合约地址变更: $NEW_ADDR (预期 $EXPECTED)"
    echo "  请更新环境变量 CONTRACT_ADDR 后重新执行 start.sh"
fi
export CONTRACT_ADDR="${NEW_ADDR:?部署未产出合约地址}"

# ---------- ⑥ 重编译后端 + 启动 ----------
echo ""
echo "[6/6] 编译后端 + 启动服务..."
cd /opt/skill-credential/backend || exit 1
export GOPROXY=https://goproxy.cn,direct
if ! go build -o skill-credential-server .; then
    echo "  后端编译失败，终止启动"
    exit 1
fi
echo "  后端编译完成"

bash /opt/skill-credential/start.sh

echo ""
echo "================================================"
echo "  环境重建完成！"
echo "  合约地址: ${NEW_ADDR}"
echo "  下一步: 通过前端录入成果并上链，创建演示存证"
echo "================================================"

