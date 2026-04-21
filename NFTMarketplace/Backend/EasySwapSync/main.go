package main

import (
	"github.com/ProjectsTask/EasySwapSync/cmd"
)

// main 是程序的入口函数。
// 当程序启动时，会首先执行这个函数。
// 它调用了 cmd 包中的 Execute 函数，用于执行命令行操作。
func main() {
	// 调用 cmd 包中的 Execute 函数，执行命令行操作
	cmd.Execute()
}
