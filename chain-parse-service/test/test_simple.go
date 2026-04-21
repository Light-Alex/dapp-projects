package main

import (
	"fmt"

	"unified-tx-parser/internal/logger"
)

func main() {
	// 创建测试日志器
	testLog := logger.New("test", "main")

	// 测试日志
	testLog.Info("这是一条Info级别的日志")
	testLog.Warn("这是一条Warn级别的日志")
	testLog.Error("这是一条Error级别的日志")

	// 测试嵌套函数
	nestedFunction()

	fmt.Println("日志测试完成")
}

func nestedFunction() {
	testLog := logger.New("test", "nested")
	testLog.Info("这是嵌套函数中的日志")
}
