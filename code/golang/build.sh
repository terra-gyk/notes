#!/bin/bash

# 定义项目根目录
PROJECT_ROOT=$(pwd)

# 编译 name_service
echo "开始编译 name_service..."
cd "$PROJECT_ROOT/name_service"
go build -o name_service

if [ $? -ne 0 ]; then
  echo "name_service 编译失败"
  exit 1
fi
echo "name_service 编译成功"

# 编译 score_service
echo "开始编译 score_service..."
cd "$PROJECT_ROOT/score_service"
go build -o score_service

if [ $? -ne 0 ]; then
  echo "score_service 编译失败"
  exit 1
fi
echo "score_service 编译成功"

echo "所有服务编译完成"