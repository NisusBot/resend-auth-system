#!/bin/bash

echo "=== Resend邮箱验证码审核系统测试 ==="
echo ""

# 启动服务器
echo "1. 启动服务器..."
cd /home/nisus/multica_workspaces/f7e3a989-c209-443f-a62b-a6ca77344594/f9bc2a48/workdir/resend-auth-system
timeout 10 go run ./cmd/server > server.log 2>&1 &
SERVER_PID=$!
sleep 3

echo "2. 测试健康检查接口..."
curl -s http://localhost:8082/api/health | jq -r '.status'
echo ""

echo "3. 测试登录接口（使用魔法验证码）..."
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8082/api/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","magic_code":"123456"}')
echo $LOGIN_RESPONSE | jq -r '.message'
TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.data.token')
echo "获取到的Token: ${TOKEN:0:20}..."

echo ""
echo "4. 测试验证码验证接口..."
VERIFY_RESPONSE=$(curl -s -X POST http://localhost:8082/api/verify \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","code":"123456"}')
echo $VERIFY_RESPONSE | jq -r '.message'
NEW_TOKEN=$(echo $VERIFY_RESPONSE | jq -r '.data.token')

echo ""
echo "5. 测试获取当前用户信息（需要认证）..."
curl -s -X GET http://localhost:8082/api/me \
  -H "Authorization: Bearer $NEW_TOKEN" | jq -r '.data.email'

echo ""
echo "6. 测试获取音乐列表（需要认证）..."
MUSIC_RESPONSE=$(curl -s -X GET http://localhost:8082/api/music \
  -H "Authorization: Bearer $NEW_TOKEN")
echo "获取到 $(echo $MUSIC_RESPONSE | jq -r '.data | length') 首音乐"

echo ""
echo "7. 测试获取系统配置..."
curl -s http://localhost:8082/api/config | jq -r '.data.ui.title'

echo ""
echo "8. 测试退出登录..."
LOGOUT_RESPONSE=$(curl -s -X POST http://localhost:8082/api/logout \
  -H "Authorization: Bearer $NEW_TOKEN")
echo $LOGOUT_RESPONSE | jq -r '.message'

# 停止服务器
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null

echo ""
echo "=== 测试完成 ==="