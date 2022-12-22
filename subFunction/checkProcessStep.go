// subFunction of vpnHelper
// Powered By Luckykeeper <luckykeeper@luckykeeper.site | https://luckykeeper.site>
package subFunction

import (
	"log"
	"os"
	"time"
)

const (
	processTagFolder = "./data/run_tags"
)

// 检查当前系统流程，对断电和意外退出的对策
func CheckProcessStep(PanelWebAddress, PanelWebPort, PanelWebUsername, PanelWebPassword, UserEmailDomain string) (status string) {
	if getConnNodeInfoTag, _ := PathExists(processTagFolder + "/getConnNodeInfoTag"); !getConnNodeInfoTag { // 从未获取过节点信息，先获取节点信息
		log.Println("检测到从未获取过节点信息，需要先获取节点信息")
		return "needGetConnNodeInfo"
	} else {
		lastTimeGetConnNodeInfoStrFile, _ := os.ReadFile(processTagFolder + "/getConnNodeInfoTag")

		lastTimeGetConnNodeInfoStr := string(lastTimeGetConnNodeInfoStrFile)
		lastTimeGetConnNodeInfo, _ := time.Parse(time.RFC822, lastTimeGetConnNodeInfoStr)

		runDuration := time.Since(lastTimeGetConnNodeInfo)

		getConnNodeInfoProcessing, _ := PathExists(processTagFolder + "/getConnNodeInfoProcessing")
		if getConnNodeInfoProcessing {
			return "getConnNodeInfoProcess"
		} else if runDuration > 24*time.Hour && !getConnNodeInfoProcessing { // 获取节点信息超过 24 H 且获取节点信息流程未在进行，获取节点信息
			log.Println("检测到获取的节点信息可能过时，需要先获取节点信息")
			return "connNodeInfoOutDate"

		} else {
			if createUserProcessing, _ := PathExists(processTagFolder + "/CreateUserProcess"); createUserProcessing {
				log.Println("检测到未完成的新增用户操作,优先完成该操作")
				return "createUserProcess"
			} else if deleteUserProcessing, _ := PathExists(processTagFolder + "/deleteUserProcessing"); deleteUserProcessing {
				log.Println("检测到未完成的删除用户操作,优先完成该操作")
				return "deleteUserProcess"
			} else {
				return "idle"
			}
		}
	}

	// 查找或创建正在运行标记
	// 准备添加到 CheckProcess，上面的 tag 也是
	// if CreateUserProcessTagIsExists, _ := PathExists(CreateUserProcessTag); !CreateUserProcessTagIsExists {
	// 	log.Println("开始执行新增用户流程，执行中标记已添加")
	// }
}
