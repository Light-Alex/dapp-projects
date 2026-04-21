package cmd

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/ProjectsTask/EasySwapBase/logger/xzap"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/ProjectsTask/EasySwapSync/service"
	"github.com/ProjectsTask/EasySwapSync/service/config"
)

// DaemonCmd 是一个 cobra 命令，用于启动同步服务守护进程。
var DaemonCmd = &cobra.Command{
	// Use 是命令的使用说明，用户可以通过这个名称来调用此命令。
	Use: "daemon",
	// Short 是命令的简短描述，用于快速了解命令的作用。
	Short: "sync easy swap order info.",
	// Long 是命令的详细描述，提供更全面的信息。
	Long: "sync easy swap order info.",
	// Run 是命令执行时调用的函数，包含了服务启动和管理的逻辑。
	Run: func(cmd *cobra.Command, args []string) {
		// 创建一个同步等待组，用于等待所有 goroutine 完成。
		wg := &sync.WaitGroup{}
		// 增加等待组的计数，表示有一个 goroutine 需要等待。
		wg.Add(1)
		// 创建一个背景上下文，作为所有后续上下文的基础。
		ctx := context.Background()
		// 创建一个可取消的上下文，用于在需要时取消所有依赖此上下文的操作。
		ctx, cancel := context.WithCancel(ctx)

		// rpc退出信号通知chan，用于接收同步服务退出时的错误信息。
		onSyncExit := make(chan error, 1)

		// 启动一个 goroutine 来执行同步服务的初始化和启动逻辑。
		go func() {
			// 当 goroutine 结束时，减少等待组的计数。
			defer wg.Done()

			// 读取和解析配置文件
			cfg, err := config.UnmarshalCmdConfig()
			if err != nil {
				// 记录配置解析失败的错误信息
				xzap.WithContext(ctx).Error("Failed to unmarshal config", zap.Error(err))
				// 将错误信息发送到退出通知通道
				onSyncExit <- err
				return
			}

			// 初始化日志模块
			_, err = xzap.SetUp(*cfg.Log)
			if err != nil {
				// 记录日志初始化失败的错误信息
				xzap.WithContext(ctx).Error("Failed to set up logger", zap.Error(err))
				// 将错误信息发送到退出通知通道
				onSyncExit <- err
				return
			}

			// 记录同步服务启动的信息，包括配置信息
			xzap.WithContext(ctx).Info("sync server start", zap.Any("config", cfg))

			// 初始化服务
			s, err := service.New(ctx, cfg)
			if err != nil {
				// 记录服务创建失败的错误信息
				xzap.WithContext(ctx).Error("Failed to create sync server", zap.Error(err))
				// 将错误信息发送到退出通知通道
				onSyncExit <- err
				return
			}

			// 启动服务
			if err := s.Start(); err != nil {
				// 记录服务启动失败的错误信息
				xzap.WithContext(ctx).Error("Failed to start sync server", zap.Error(err))
				// 将错误信息发送到退出通知通道
				onSyncExit <- err
				return
			}

			// 开启pprof，用于性能监控
			if cfg.Monitor.PprofEnable {
				// 启动一个 HTTP 服务器，监听指定端口，用于 pprof 性能监控
				http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", cfg.Monitor.PprofPort), nil)
			}
		}()

		// 信号通知chan，用于接收系统信号，如 SIGINT 和 SIGTERM。
		onSignal := make(chan os.Signal)
		// 优雅退出，监听系统信号，当接收到指定信号时执行相应操作
		signal.Notify(onSignal, syscall.SIGINT, syscall.SIGTERM)
		// 等待信号或同步服务退出通知
		select {
		case sig := <-onSignal:
			// 根据接收到的信号类型执行相应操作
			switch sig {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM:
				// 取消上下文，通知所有依赖此上下文的操作停止
				cancel()
				// 记录因信号退出的信息
				xzap.WithContext(ctx).Info("Exit by signal", zap.String("signal", sig.String()))
			}
		case err := <-onSyncExit:
			// 取消上下文，通知所有依赖此上下文的操作停止
			cancel()
			// 记录因错误退出的信息
			xzap.WithContext(ctx).Error("Exit by error", zap.Error(err))
		}
		// 等待所有 goroutine 完成
		wg.Wait()
	},
}

// init 是 Go 语言中的特殊初始化函数，在包被导入时自动执行。
func init() {
	// 将 DaemonCmd 命令添加到 rootCmd 主命令中，
	// 这样用户可以通过 rootCmd 调用 DaemonCmd 启动同步服务守护进程。
	rootCmd.AddCommand(DaemonCmd)
}
