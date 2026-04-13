#!/bin/sh

# InfluxDB多链bucket初始化脚本
# 该脚本在InfluxDB容器启动后创建各个链专用的bucket
# 注意：使用 POSIX sh 兼容语法

set -e

echo "🚀 开始初始化InfluxDB多链bucket..."

# InfluxDB配置
INFLUX_TOKEN="${DOCKER_INFLUXDB_INIT_ADMIN_TOKEN}"
INFLUX_ORG="${DOCKER_INFLUXDB_INIT_ORG}"

# 步骤1: 等待InfluxDB初始化完成
echo "⏳ 等待InfluxDB初始化..."
MAX_WAIT=120
WAIT_COUNT=0
while [ $WAIT_COUNT -lt $MAX_WAIT ]; do
  if /usr/local/bin/influx ping > /dev/null 2>&1; then
    echo "✅ InfluxDB已就绪"
    break
  fi
  WAIT_COUNT=$((WAIT_COUNT + 1))
  sleep 1
done

if [ $WAIT_COUNT -eq $MAX_WAIT ]; then
  echo "❌ 等待InfluxDB超时"
  exit 1
fi

# 额外等待确保服务完全就绪
sleep 5

# 步骤2: 创建各个链的专用bucket
echo "📦 开始创建bucket..."
RETENTION="90d"

for bucket in sui ethereum bsc solana; do
  echo "📦 处理bucket: $bucket"

  # 检查bucket是否已存在
  if /usr/local/bin/influx bucket list \
    --token $INFLUX_TOKEN \
    --org $INFLUX_ORG \
    --name $bucket > /dev/null 2>&1; then
    echo "⚠️  Bucket '$bucket' 已存在，跳过创建"
  else
    # 创建bucket
    if /usr/local/bin/influx bucket create \
      --token $INFLUX_TOKEN \
      --org $INFLUX_ORG \
      --name $bucket \
      --retention $RETENTION > /dev/null 2>&1; then
      echo "✅ Bucket '$bucket' 创建成功"
    else
      echo "❌ Bucket '$bucket' 创建失败"
    fi
  fi
done

echo "🎉 InfluxDB多链bucket初始化完成！"
echo "📋 已创建的bucket列表："
/usr/local/bin/influx bucket list --token $INFLUX_TOKEN --org $INFLUX_ORG
