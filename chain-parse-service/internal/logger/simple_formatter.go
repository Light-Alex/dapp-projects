package logger

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

// SimpleFormatter 简单的日志格式化器，显示相对路径
type SimpleFormatter struct {
	logrus.TextFormatter
}

// NewSimpleFormatter 创建新的简单格式化器
func NewSimpleFormatter() *SimpleFormatter {
	return &SimpleFormatter{
		TextFormatter: logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		},
	}
}

// Format 格式化日志条目
func (f *SimpleFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// 获取调用者信息
	if entry.Caller != nil {
		// 获取文件名和行号
		filename := entry.Caller.File
		line := entry.Caller.Line
		functionName := entry.Caller.Function

		// 提取函数名（去掉包路径）
		lastSlash := strings.LastIndex(functionName, "/")
		if lastSlash > 0 {
			functionName = functionName[lastSlash+1:]
		}
		lastDot := strings.LastIndex(functionName, ".")
		if lastDot > 0 {
			functionName = functionName[lastDot+1:]
		}

		// 获取相对路径（去掉项目根目录前缀）
		// 动态获取当前工作目录作为项目根目录
		projectRoot, _ := os.Getwd()
		relPath := strings.TrimPrefix(filename, projectRoot+"/")
		// 如果没有匹配到工作目录，尝试使用原始路径
		if relPath == filename {
			// 尝试提取最后的文件名部分
			lastSlash := strings.LastIndex(filename, "/")
			if lastSlash > 0 {
				relPath = filename[lastSlash+1:]
			}
		}

		// 创建新的消息格式
		location := relPath + ":" + fmt.Sprintf("%d", line) + " (" + functionName + ")"
		entry.Message = "[" + location + "] " + entry.Message
	}

	// 使用TextFormatter的默认格式，但禁用调用者信息
	f.TextFormatter.CallerPrettyfier = func(frame *runtime.Frame) (function string, file string) {
		return "", ""
	}

	return f.TextFormatter.Format(entry)
}
