// vpnHelper
// Powered By Luckykeeper <luckykeeper@luckykeeper.site | https://luckykeeper.site>

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"github.com/urfave/cli/v2"
	"gopkg.in/ini.v1"

	vpnHelperRouter "vpnHelper/router"
	subFunction "vpnHelper/subFunction"
)

var (
	// 面板 WEB 参数
	PanelWebAddress, PanelWebPort, PanelWebUsername, PanelWebPassword string

	// 本服务参数
	VpnHelperWebPort, UserEmailDomain, VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName string
)

// 程序入口
func main() {
	readConfig()
	subFunction.InitializeDatabase(VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName)

	// 移除意外退出导致遗留的过时 Tag
	if taskDispatcherTag, _ := subFunction.PathExists("./data/run_tags/TaskDispatcher"); taskDispatcherTag {
		os.Remove("./data/run_tags/TaskDispatcher")
	}
	if taskDispatcherTag, _ := subFunction.PathExists("./data/run_tags/getConnNodeInfoProcessing"); taskDispatcherTag {
		os.Remove("./data/run_tags/getConnNodeInfoProcessing")
	}
	if taskDispatcherTag, _ := subFunction.PathExists("./data/run_tags/CreateUserProcess"); taskDispatcherTag {
		os.Remove("./data/run_tags/CreateUserProcess")
	}
	if taskDispatcherTag, _ := subFunction.PathExists("./data/run_tags/deleteUserProcessing"); taskDispatcherTag {
		os.Remove("./data/run_tags/deleteUserProcessing")
	}

	vpnHelperCLI()
}

// CLI
func vpnHelperCLI() {
	vpnHelper := &cli.App{
		Name: "vpnHelper",
		Usage: "vpnHelper" +
			"\nPowered By Luckykeeper <luckykeeper@luckykeeper.site | https://luckykeeper.site>" +
			"\n————————————————————————————————————————" +
			"\n注意：使用前需要先填写同目录下 config.ini !",
		Version: "1.0.0_build20221214",
		Commands: []*cli.Command{
			{
				Name:    "runProd",
				Aliases: []string{"r"},
				Usage:   "启动 API 服务（生产环境）",
				Action: func(cCtx *cli.Context) error {
					apiServer(false)
					return nil
				},
			},
			{
				Name:    "runDebug",
				Aliases: []string{"debug"},
				Usage:   "启动 API 服务（Debug）",
				Action: func(cCtx *cli.Context) error {
					apiServer(true)
					return nil
				},
			},
			{
				Name:    "generateToken",
				Usage:   "新增授权 Token",
				Aliases: []string{"gt"},
				Action: func(cCtx *cli.Context) error {
					fmt.Println("正在生成 Token ……")
					subFunction.GenerateToken(VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName)
					return nil
				},
			},
			{
				Name:    "showToken",
				Usage:   "显示当前系统可用 Token",
				Aliases: []string{"st"},
				Action: func(cCtx *cli.Context) error {
					subFunction.ShowToken(VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName)
					return nil
				},
			},
			{
				Name:    "deleteToken",
				Usage:   "删除指定 Token ，需要带参数 | e.g. ./vpnHelper dt 123456",
				Aliases: []string{"dt"},
				Action: func(cCtx *cli.Context) error {
					subFunction.DeleteToken(cCtx.Args().First(), VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName)
					return nil
				},
			},
			// 临时调试功能
			// {
			// 	Name:    "getConnNodeInfo",
			// 	Usage:   "调试-获取并生成邀请码",
			// 	Aliases: []string{"gc"},
			// 	Action: func(cCtx *cli.Context) error {
			// 		subFunction.GetConnNodeInfo(PanelWebAddress, PanelWebPort, PanelWebUsername, PanelWebPassword,
			// 			VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName)
			// 		return nil
			// 	},
			// },
		},
	}

	if err := vpnHelper.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

// read config.ini
func readConfig() {
	configFile, err := ini.Load("config.ini")
	if err != nil {
		fmt.Printf("读取配置文件 config.ini 失败，失败原因为: %v", err)
		os.Exit(1)
	}
	PanelWebAddress = configFile.Section("panelWeb").Key("address").String()
	PanelWebPort = configFile.Section("panelWeb").Key("port").String()
	PanelWebUsername = configFile.Section("panelWeb").Key("username").String()
	PanelWebPassword = configFile.Section("panelWeb").Key("password").String()

	VpnHelperWebPort = configFile.Section("vpnHelper").Key("webPort").String()
	UserEmailDomain = configFile.Section("vpnHelper").Key("userEmailDomain").String()
	VpnHelperDBAddress = configFile.Section("vpnHelper").Key("dbAddress").String()
	VpnHelperDBPort = configFile.Section("vpnHelper").Key("dbPort").String()
	VpnHelperDBUsername = configFile.Section("vpnHelper").Key("dbUsername").String()
	VpnHelperDBPassword = configFile.Section("vpnHelper").Key("dbPassword").String()
	VpnHelperDBName = configFile.Section("vpnHelper").Key("dbName").String()
}

// vpnHelper - runProd,runDebug
func apiServer(apiServerDebug bool) {
	if !apiServerDebug {
		// 生产环境
		gin.SetMode(gin.ReleaseMode)
	} else {
		// Debug
		logSystem()
	}
	router := gin.Default()
	vpnHelperRouter.GinRouter(router)

	// 内部定时任务对象
	c := cron.New(cron.WithSeconds())
	// 给对象增加定时任务
	c.AddFunc("* */5 * * * *", func() {
		subFunction.TaskDispatcher(PanelWebAddress, PanelWebPort, PanelWebUsername, PanelWebPassword, UserEmailDomain,
			VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName)
	})
	c.Start()

	log.Println("API 服务已启动，端口:", VpnHelperWebPort)
	router.Run(":" + VpnHelperWebPort)
}

// 日志记录
func logSystem() {
	logfileAddress := "./data/log/vpnHelper.log"
	if logfileExists, _ := subFunction.PathExists(logfileAddress); logfileExists {
		os.Remove(logfileAddress)
		log.Println("删除已经存在的日志文件！")
	} else {
		log.Println("日志文件不存在，在" + logfileAddress + "创建日志并开始记录")
	}
	logfile, err := os.OpenFile(logfileAddress, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0777)

	if err != nil {
		os.Exit(1)
	}

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(logfile)
}
