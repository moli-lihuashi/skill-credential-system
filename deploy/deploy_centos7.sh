#!/bin/bash
# ============================================================
# 基于联盟链的高校毕业生技能成果可信存证系统
# CentOS7 一键部署脚本
# 使用方法: bash deploy_centos7.sh
#
# 依赖环境变量（禁止把口令/私钥写进本文件）：
#   MYSQL_ROOT_PW   MySQL root 口令
#   DB_PASSWORD     应用数据库口令
#   DEPLOYER_KEY    合约部署者私钥
#   JWT_SECRET      JWT 签名密钥
# ============================================================

set -e

PROJECT_DIR="/opt/skill-credential"
DB_NAME="skill_credential"
DB_APP_USER="skill_app"

: "${MYSQL_ROOT_PW:?请通过环境变量提供 MYSQL_ROOT_PW}"
: "${DB_PASSWORD:?请通过环境变量提供 DB_PASSWORD}"
: "${DEPLOYER_KEY:?请通过环境变量提供 DEPLOYER_KEY}"
: "${JWT_SECRET:?请通过环境变量提供 JWT_SECRET}"

echo "=========================================="
echo " 技能成果可信存证系统 - CentOS7 部署脚本"
echo "=========================================="

# ---------- 1. 安装基础依赖 ----------
echo "[1/7] 安装基础依赖..."
yum install -y wget git gcc gcc-c++ make vim

# ---------- 2. 安装 MySQL 5.7 ----------
echo "[2/7] 安装 MySQL..."
if ! command -v mysql &>/dev/null; then
    rpm -Uvh https://dev.mysql.com/get/mysql57-community-release-el7-11.noarch.rpm
    yum install -y mysql-community-server
    systemctl start mysqld
    systemctl enable mysqld

    # 获取临时密码并修改
    TEMP_PW=$(grep 'temporary password' /var/log/mysqld.log | awk '{print $NF}')
    mysql --connect-expired-password -uroot -p"${TEMP_PW}" <<EOF
ALTER USER 'root'@'localhost' IDENTIFIED BY '${MYSQL_ROOT_PW}FL@2024';
EOF
    mysql -uroot -p"${MYSQL_ROOT_PW}FL@2024" <<EOF
ALTER USER 'root'@'localhost' IDENTIFIED BY '${MYSQL_ROOT_PW}';
CREATE DATABASE IF NOT EXISTS ${DB_NAME} DEFAULT CHARACTER SET utf8mb4;
-- 最小权限应用账号：仅本机访问、仅本库权限，禁止 root@'%'
CREATE USER IF NOT EXISTS '${DB_APP_USER}'@'localhost' IDENTIFIED BY '${DB_PASSWORD}';
GRANT ALL PRIVILEGES ON ${DB_NAME}.* TO '${DB_APP_USER}'@'localhost';
FLUSH PRIVILEGES;
EOF
else
    echo "MySQL 已安装，跳过"
fi

# ---------- 3. 安装 Go ----------
echo "[3/7] 安装 Go..."
if ! command -v go &>/dev/null; then
    cd /tmp
    wget -q https://golang.google.cn/dl/go1.21.10.linux-amd64.tar.gz
    tar -C /usr/local -xzf go1.21.10.linux-amd64.tar.gz
    echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
    export PATH=$PATH:/usr/local/go/bin
    go version
else
    echo "Go 已安装，跳过"
fi

# ---------- 4. 安装 Node.js ----------
echo "[4/7] 安装 Node.js..."
if ! command -v node &>/dev/null; then
    curl -fsSL https://rpm.nodesource.com/setup_18.x | bash -
    yum install -y nodejs
fi

# ---------- 5. 安装 Ganache（本地联盟链节点） ----------
echo "[5/7] 安装 Ganache 本地链..."
npm install -g ganache 2>/dev/null || echo "Ganache 安装可能需要网络"

# ---------- 6. 部署项目 ----------
echo "[6/7] 部署项目文件..."
mkdir -p ${PROJECT_DIR}
# 确保源码/构建产物位于 ${PROJECT_DIR}，与 Nginx root 保持一致
# 若当前不在项目目录，则要求用户先拷贝源码再执行
SCRIPT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
if [ ! -d "${SCRIPT_DIR}/backend" ] || [ ! -d "${SCRIPT_DIR}/frontend" ]; then
    echo "错误: 未在项目根目录执行，且脚本附近找不到 backend/frontend"
    exit 1
fi
cp -a "${SCRIPT_DIR}/." "${PROJECT_DIR}/"
cd "${PROJECT_DIR}"

# 初始化数据库
mysql -uroot -p"${MYSQL_ROOT_PW}" < sql/init.sql
echo "数据库初始化完成"

# 构建后端
cd backend
go mod tidy
go build -o skill-credential-server .
echo "后端构建完成"

# 构建前端
cd ../frontend
npm install
npm run build
echo "前端构建完成"

# ---------- 7. 配置 Nginx ----------
echo "[7/7] 配置 Nginx..."
# CentOS7 的 nginx 在 EPEL 源中
yum install -y epel-release || true
yum install -y nginx

cat > /etc/nginx/conf.d/skill-credential.conf <<'NGINX'
server {
    listen 80;
    server_name _;

    root /opt/skill-credential/frontend/dist;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
NGINX

systemctl start nginx
systemctl enable nginx


# ---------- 启动服务 ----------
echo "=========================================="
echo " 部署完成！"
echo "=========================================="
echo ""
echo "启动步骤:"
echo "  1. 启动链节点:  mkdir -p /opt/skill-credential/chain-data && ganache --port 8545 --chainId 1337 --db /opt/skill-credential/chain-data"
echo "  2. 部署合约:    cd contracts && node deploy.js   # 需先 export DEPLOYER_KEY=..."
echo "  3. 启动后端:    cd backend && CONTRACT_ADDR=<合约地址> JWT_SECRET=<密钥> DB_PASSWORD=<口令> ./skill-credential-server"
echo "  4. 访问前端:    http://<虚拟机IP>/"
echo ""
echo "MySQL 应用账号: ${DB_APP_USER}@localhost（口令通过 DB_PASSWORD 注入，不在此回显）"
echo "项目目录: ${PROJECT_DIR}"

