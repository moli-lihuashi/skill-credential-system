#!/bin/bash
# ============================================================
# 技能成果可信存证系统 - 一键启动
# ============================================================
export PATH=/usr/local/go/bin:/usr/local/nodejs/bin:$PATH
export LANG=zh_CN.UTF-8
export LC_ALL=zh_CN.UTF-8
timedatectl set-timezone Asia/Shanghai 2>/dev/null

# 敏感项（数据库口令 / 部署者私钥 / JWT 密钥）不写入源码：
# 优先取环境变量，未设置时回退到同目录下的 .env.demo（本地文件，已被 .gitignore 忽略）
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
if [ -f "$SCRIPT_DIR/.env.demo" ]; then
    set -a
    # shellcheck disable=SC1091
    . "$SCRIPT_DIR/.env.demo"
    set +a
fi
: "${DB_PASSWORD:?请通过环境变量或 deploy/.env.demo 提供 DB_PASSWORD}"
: "${DEPLOYER_KEY:?请通过环境变量或 deploy/.env.demo 提供 DEPLOYER_KEY}"
: "${JWT_SECRET:?请通过环境变量或 deploy/.env.demo 提供 JWT_SECRET}"
export DB_PASSWORD DEPLOYER_KEY JWT_SECRET

# 合约地址非敏感项，保留演示默认值
export CONTRACT_ADDR="${CONTRACT_ADDR:-0xe78A0F7E598Cc8b0Bb87894B0F60dD2a88d6a8Ab}"

echo "================================================"
echo "  技能成果可信存证系统 - 一键启动"
echo "================================================"

# 1. MySQL
echo ""
echo "[1/5] 启动 MySQL..."
systemctl start mysqld 2>/dev/null
sleep 2
if mysql -uroot -p"$DB_PASSWORD" -e "SELECT 1;" 2>/dev/null | grep -q "1"; then
    echo "  ✅ MySQL OK"
else
    echo "  ❌ MySQL FAIL"
fi

# 2. Ganache
echo ""
echo "[2/5] 启动 Ganache..."
kill $(ps aux | grep ganache | grep -v grep | awk '{print $2}') 2>/dev/null
sleep 1
nohup ganache --host 0.0.0.0 --port 8545 -i 1337 -a 10 -e 1000 --wallet.deterministic --db /opt/skill-credential/chain-data > /tmp/ganache.log 2>&1 &
sleep 5
if curl -s -X POST -H 'Content-Type: application/json' --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' http://127.0.0.1:8545 | grep -q "result"; then
    echo "  ✅ Ganache OK (8545)"
else
    echo "  ❌ Ganache FAIL"
fi

# 3. Backend
echo ""
echo "[3/5] 启动 Backend..."
kill $(ps aux | grep skill-credential-server | grep -v grep | awk '{print $2}') 2>/dev/null
sleep 1
export CHAIN_RPC=http://127.0.0.1:8545
export UPLOAD_DIR=/opt/skill-credential/uploads
mkdir -p /opt/skill-credential/uploads
cd /opt/skill-credential/backend
nohup ./skill-credential-server > /tmp/backend.log 2>&1 &
sleep 4
if curl -s http://127.0.0.1:8080/health | grep -q "ok"; then
    echo "  ✅ Backend OK (8080)"
else
    echo "  ❌ Backend FAIL"
fi

# 4. Nginx
echo ""
echo "[4/5] 启动 Nginx..."
find /opt/skill-credential/frontend/dist -type f -exec chmod 644 {} \; 2>/dev/null
find /opt/skill-credential/frontend/dist -type d -exec chmod 755 {} \; 2>/dev/null
setenforce 0 2>/dev/null
systemctl start nginx 2>/dev/null
nginx -s reload 2>/dev/null
if curl -s http://127.0.0.1/ | grep -q "index-"; then
    echo "  ✅ Nginx OK (80)"
else
    echo "  ❌ Nginx FAIL"
fi

# 5. Verify
echo ""
echo "[5/5] 系统验证..."
echo "  Health:   $(curl -s http://127.0.0.1:8080/health)"
echo "  Frontend: $(curl -s http://127.0.0.1/ | grep -o 'index-[A-Za-z0-9_-]*\.js')"

echo ""
echo "================================================"
echo "  全部服务启动完成！"
echo ""
echo "  http://<虚拟机IP>"
echo ""
echo "  测试账号（口令不在脚本中明文给出，见本地 deploy/.env.demo）:"
echo "    学生:      demo_student"
echo "    院校管理员: admin01"
echo "    实训企业:  company01"
echo "    招聘企业:  employer01"
echo "================================================"
