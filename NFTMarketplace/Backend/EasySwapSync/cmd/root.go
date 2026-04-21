package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/mitchellh/go-homedir" //获取用户主目录路径的库
	"github.com/spf13/cobra"          //用于创建强大的CLI应用程序的命令行库，支持命令、参数、标志等功能。
	"github.com/spf13/viper"          //一个配置管理库，能够处理JSON、TOML、YAML、HCL、INI、envfile和Java properties格式的文件，并支持从环境变量、命令行参数、远程配置系统等多个来源获取配置。
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "sync",
	Short: "root server.",
	Long:  `root server.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
// Execute 函数用于执行根命令，并处理可能出现的错误。
// 该函数会在执行根命令失败时退出程序，并打印配置文件路径。
func Execute() {
	// 执行根命令
	if err := rootCmd.Execute(); err != nil {
		// 如果执行出错，打印错误信息
		fmt.Println(err)
		// 退出程序，返回状态码1
		os.Exit(1)
	}

	// 打印配置文件路径
	fmt.Println("cfgFile=", cfgFile)
}

// init 是 Go 语言中的特殊初始化函数，在包被导入时自动执行。
// 此函数用于设置命令行标志和初始化配置。
func init() {
	// 设置 initConfig 函数在调用 rootCmd 的 Execute() 方法时运行，
	// 确保在执行命令之前加载配置文件。
	cobra.OnInitialize(initConfig)

	// 获取 rootCmd 的持久标志集，持久标志可用于该命令及其所有子命令。
	flags := rootCmd.PersistentFlags()

	// 定义一个名为 "config" 的字符串类型标志，简称 "c"。
	// 该标志的值将存储在 cfgFile 变量中。
	// 默认值为 "./config/config_import.toml"，并提供了一个说明信息。
	flags.StringVarP(&cfgFile, "config", "c", "./config/config_import.toml", "config file (default is $HOME/.config_import.toml)")
}

// initConfig reads in config file and ENV variables if set.
// initConfig 函数用于读取配置文件和环境变量。
// 如果指定了配置文件路径，则使用该文件；否则，在用户主目录下搜索配置文件。
func initConfig() {
	// 检查是否通过命令行标志指定了配置文件
	if cfgFile != "" {
		// 从命令行标志中获取配置文件路径，并设置给 viper
		viper.SetConfigFile(cfgFile)
	} else {
		// 获取当前用户的主目录路径
		home, err := homedir.Dir()
		if err != nil {
			// 若获取主目录失败，打印错误信息并退出程序
			fmt.Println(err)
			os.Exit(1)
		}

		// 向 viper 添加主目录作为配置文件搜索路径
		viper.AddConfigPath(home)
		// 设置要搜索的配置文件的名称（不包含扩展名）
		viper.SetConfigName("config_import")
	}

	// 让 viper 自动读取环境变量
	viper.AutomaticEnv()
	// 设置配置文件的类型为 toml
	viper.SetConfigType("toml")
	// 设置环境变量的前缀，以区分不同应用的环境变量
	viper.SetEnvPrefix("EasySwap")
	// 创建一个字符串替换器，将配置文件中的 "." 替换为 "_"，以适应环境变量的命名规则
	replacer := strings.NewReplacer(".", "_")
	// 设置 viper 在读取环境变量时使用替换器
	viper.SetEnvKeyReplacer(replacer)

	// 尝试读取配置文件
	if err := viper.ReadInConfig(); err == nil {
		// 若读取成功，打印正在使用的配置文件路径
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		// 若读取失败，抛出错误并终止程序
		panic(err)
	}
}
