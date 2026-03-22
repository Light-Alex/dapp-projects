package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"blockchain-parser-go/config"
	"blockchain-parser-go/database"
	"blockchain-parser-go/parser"
	"blockchain-parser-go/redis"
	"blockchain-parser-go/service"
)

func main() {
	// 显示短文件名和行号
	log.SetFlags(log.Lshortfile)

	// Load configuration
	cfg := config.LoadConfig()

	log.Printf("BSC_RPC_URL: %s", cfg.RPCURL)

	// Initialize database
	db, err := database.NewDB(
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
	)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize Redis
	redisClient, err := redis.NewRedisClient(
		cfg.RedisHost,
		cfg.RedisPort,
		cfg.RedisPassword,
	)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	// Initialize blockchain parser
	blockchainParser, err := parser.NewBlockchainParser(cfg, db, redisClient)
	if err != nil {
		log.Fatalf("Failed to initialize blockchain parser: %v", err)
	}

	// Initialize withdrawal service
	withdrawalService, err := service.NewWithdrawalService(cfg, db)
	if err != nil {
		log.Fatalf("Failed to initialize newWithdrawal: %v", err)
	}

	blockchainParser.InitData()
	blockchainParser.Start()
	withdrawalService.Start()

	// 等待 Ctrl+C、Kill 信号，结束主进程
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down...")
	os.Exit(0)
}
