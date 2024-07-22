package main

import (
	"best-ticker/config"
	"best-ticker/context"
	"fmt"
	_ "net/http"
	_ "net/http/pprof"

	"best-ticker/utils"
	"best-ticker/utils/logger"
	"os"
	"runtime"
	"time"
)

var globalConfig config.Config
var globalContext context.GlobalContext

func ExitProcess() {
	// 停止挂单
	logger.Info("[Exit] stop program.")
	os.Exit(1)
}

func main() {
	runtime.GOMAXPROCS(1)
	// 参数判断
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s config_file\n", os.Args[0])
		os.Exit(1)
	}

	// 监听退出消息，并调用ExitProcess进行处理
	utils.RegisterExitSignal(ExitProcess)

	// 加载配置文件
	globalConfig = *config.LoadConfig(os.Args[1])

	// 设置日志级别, 并初始化日志
	logger.InitLogger(globalConfig.LogPath, globalConfig.LogLevel)

	// 解析config，加载杠杆和合约交易对，初始化context，账户初始化设置，拉取仓位、余额等
	globalContext.Init(&globalConfig)

	// 启动zmq服务
	StartZmq()

	// 开始监听ticker消息
	startTickerMessage()

	// 开始监听orderBook消息
	startDepthMessage()

	// 阻塞主进程
	for {
		time.Sleep(24 * time.Hour)
	}
}
