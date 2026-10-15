#!/bin/bash
# ============================================================
# 功能测试 + 篡改模拟测试脚本（修正版）
# 前提：后端已启动 http://127.0.0.1:8080
# ============================================================

BASE="http://127.0.0.1:8080/api"
PASS=0
FAIL=0

# 演示账号口令通过环境变量注入，不写入脚本
: "${DEMO_PASSWORD:?请通过环境变量提供 DEMO_PASSWORD（见 deploy/.env.demo）}"

check() {
    local desc=$1
    local result=$2
    if [ "$result" = "true" ]; then
        echo "  [PASS] $desc"
        PASS=$((PASS+1))
    else
        echo "  [FAIL] $desc"
        FAIL=$((FAIL+1))
    fi
}

echo "=========================================="
echo "  功能测试开始"
echo "=========================================="

# ---------- 1. 健康检查 ----------
echo ""
echo "1. 系统健康检查"
HEALTH=$(curl -s http://127.0.0.1:8080/health)
echo "   $HEALTH"
check "后端健康检查" "$(echo "$HEALTH" | grep -q 'ok' && echo true || echo false)"

# ---------- 2. 登录获取 Token（使用实际账号） ----------
echo ""
echo "2. 登录测试"
LOGIN=$(curl -s -X POST $BASE/auth/login -H "Content-Type: application/json" -d "{
    \"username\": \"admin01\",
    \"password\": \"${DEMO_PASSWORD}\"
}")
TOKEN=$(echo $LOGIN | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
check "管理员登录" "$([ -n "$TOKEN" ] && echo true || echo false)"

# ---------- 3. 学生登录 ----------
echo ""
echo "3. 学生登录"
SLOGIN=$(curl -s -X POST $BASE/auth/login -H "Content-Type: application/json" -d "{
    \"username\": \"demo_student\",
    \"password\": \"${DEMO_PASSWORD}\"
}")
STOKEN=$(echo $SLOGIN | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
check "学生登录" "$([ -n "$STOKEN" ] && echo true || echo false)"

# ---------- 4. 文件上传与哈希计算 ----------
echo ""
echo "4. 文件上传与哈希计算"
echo "Test certificate content $(date)" > /tmp/test_file.txt
UPLOAD=$(curl -s -X POST $BASE/upload -H "Authorization: Bearer $TOKEN" -F "file=@/tmp/test_file.txt")
FILE_HASH=$(echo $UPLOAD | grep -o '"file_hash":"[^"]*"' | cut -d'"' -f4)
check "上传文件并计算哈希" "$([ -n "$FILE_HASH" ] && echo true || echo false)"
echo "   文件哈希: $FILE_HASH"

# ---------- 5. 存证查询 ----------
echo ""
echo "5. 查询已有存证"
CRED_LIST=$(curl -s $BASE/credential/my -H "Authorization: Bearer $STOKEN")
CRED_NO=$(echo $CRED_LIST | grep -o '"credential_no":"[^"]*"' | head -1 | cut -d'"' -f4)
CHAIN_HASH=$(echo $CRED_LIST | grep -o '"file_hash":"[^"]*"' | head -1 | cut -d'"' -f4)
check "查询存证列表" "$([ -n "$CRED_NO" ] && echo true || echo false)"
echo "   存证编号: $CRED_NO"
echo "   链上哈希: $CHAIN_HASH"

# ---------- 6. 核验测试：正确哈希 ----------
echo ""
echo "6. 核验测试（正确哈希，应通过）"
if [ -n "$CRED_NO" ] && [ -n "$CHAIN_HASH" ]; then
    VERIFY_PASS=$(curl -s -X POST $BASE/verify -H "Content-Type: application/json" -d "{
        \"credential_no\": \"$CRED_NO\",
        \"file_hash\": \"$CHAIN_HASH\"
    }")
    RESULT=$(echo $VERIFY_PASS | grep -o '"result":"[^"]*"' | cut -d'"' -f4)
    echo "   核验结果: $VERIFY_PASS" | head -c 200
    echo ""
    check "正确哈希核验通过" "$([ "$RESULT" = "pass" ] && echo true || echo false)"
else
    check "正确哈希核验通过" "false"
fi

# ---------- 7. 篡改模拟测试 ----------
echo ""
echo "7. 篡改模拟测试（改哈希，应不通过）"
if [ -n "$CRED_NO" ]; then
    # 篡改哈希最后一位
    TAMPERED_HASH="${CHAIN_HASH%?}9"
    if [ "$TAMPERED_HASH" = "$CHAIN_HASH" ]; then
        TAMPERED_HASH="${CHAIN_HASH%?}0"
    fi
    echo "   原始哈希: $CHAIN_HASH"
    echo "   篡改哈希: $TAMPERED_HASH"
    VERIFY_FAIL=$(curl -s -X POST $BASE/verify -H "Content-Type: application/json" -d "{
        \"credential_no\": \"$CRED_NO\",
        \"file_hash\": \"$TAMPERED_HASH\"
    }")
    RESULT=$(echo $VERIFY_FAIL | grep -o '"result":"[^"]*"' | cut -d'"' -f4)
    echo "   核验结果: $VERIFY_FAIL" | head -c 200
    echo ""
    check "篡改哈希核验不通过" "$([ "$RESULT" = "fail" ] && echo true || echo false)"
else
    check "篡改哈希核验不通过" "false"
fi

# ---------- 8. 并发核验测试 ----------
echo ""
echo "8. 并发核验测试（10 个并发）"
if [ -n "$CRED_NO" ] && [ -n "$CHAIN_HASH" ]; then
    # 跑前记录 verify_count
    BEFORE=$(curl -s $BASE/credential/my -H "Authorization: Bearer $STOKEN")
    VC_BEFORE=$(echo "$BEFORE" | grep -o '"verify_count":[0-9]*' | head -1 | cut -d: -f2)
    echo "   跑前核验次数: $VC_BEFORE"

    # 10 个并发请求
    for i in {1..10}; do
        curl -s -X POST $BASE/verify -H "Content-Type: application/json" -d "{
            \"credential_no\": \"$CRED_NO\",
            \"file_hash\": \"$CHAIN_HASH\"
        }" > /dev/null &
    done
    wait

    # 跑后记录 verify_count
    AFTER=$(curl -s $BASE/credential/my -H "Authorization: Bearer $STOKEN")
    VC_AFTER=$(echo "$AFTER" | grep -o '"verify_count":[0-9]*' | head -1 | cut -d: -f2)
    echo "   跑后核验次数: $VC_AFTER"
    DIFF=$((VC_AFTER - VC_BEFORE))
    echo "   增量: $DIFF (期望 >= 10)"
    check "并发核验 10 次完成" "$([ "$DIFF" -ge 10 ] && echo true || echo false)"
else
    check "并发核验 10 次完成" "false"
fi

# ---------- 9. 日志上链检查 ----------
echo ""
echo "9. 日志上链检查"
LOGS=$(curl -s $BASE/logs -H "Authorization: Bearer $TOKEN")
ONCHAIN=$(echo $LOGS | grep -o '"onchain":1' | wc -l)
TOTAL=$(echo $LOGS | grep -o '"onchain":' | wc -l)
echo "   日志总数: $TOTAL, 已上链: $ONCHAIN"
check "日志哈希已上链" "$([ "$ONCHAIN" -gt 0 ] && echo true || echo false)"

# ---------- 汇总 ----------
echo ""
echo "=========================================="
echo "  测试结果: PASS=$PASS  FAIL=$FAIL"
echo "=========================================="

[ $FAIL -eq 0 ] && echo "  ✅ 全部测试通过!" || echo "  ❌ 有测试失败"
exit $FAIL
