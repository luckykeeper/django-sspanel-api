// subFunction of vpnHelper
// Powered By Luckykeeper <luckykeeper@luckykeeper.site | https://luckykeeper.site>
package subFunction

import (
	"log"
	"os"
	"time"
)

// 任务调度器，定时批量执行任务
func TaskDispatcher(PanelWebAddress, PanelWebPort, PanelWebUsername, PanelWebPassword, UserEmailDomain,
	VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName string) {
	// 先停一下等下面时间戳创建完成，因为这里的 Cron 会反复运行多个线程，后面根据时间戳具体判断当前状态
	time.Sleep(2 * time.Second)
	log.Println("定时任务 - 执行任务调度 - 判定任务分配")
	status := CheckProcessStep(PanelWebAddress, PanelWebPort, PanelWebUsername, PanelWebPassword, UserEmailDomain)
	if taskDispatcherTag, _ := PathExists(processTagFolder + "/TaskDispatcher"); !taskDispatcherTag && status == "connNodeInfoOutDate" {
		os.Create(processTagFolder + "/TaskDispatcher")
		log.Println("定时任务 - 执行任务调度 - 更新节点信息")
		GetConnNodeInfo(PanelWebAddress, PanelWebPort, PanelWebUsername, PanelWebPassword,
			VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName)
		os.Remove(processTagFolder + "/TaskDispatcher")
	} else if taskDispatcherTag, _ := PathExists(processTagFolder + "/TaskDispatcher"); !taskDispatcherTag || status == "idle" {
		os.Create(processTagFolder + "/TaskDispatcher")
		// 先检查是否有执行 Tag 存在，如果没有再定时执行
		if status == "needGetConnNodeInfo" { // 最高执行优先级：获取节点信息
			log.Println("定时任务 - 执行任务调度 - 获取节点信息")
			GetConnNodeInfo(PanelWebAddress, PanelWebPort, PanelWebUsername, PanelWebPassword,
				VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName)
		} else if status == "createUserProcess" {
			log.Println("定时任务 - 执行任务调度 - 恢复未完成的新增用户操作")
			CreateUserProcess(PanelWebAddress, PanelWebPort, PanelWebUsername, PanelWebPassword, UserEmailDomain,
				VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName)
		} else if status == "deleteUserProcess" {
			log.Println("定时任务 - 执行任务调度 - 恢复未完成的删除用户操作")
			DeleteUserProcess(PanelWebAddress, PanelWebPort, PanelWebUsername, PanelWebPassword,
				VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName,
				UserEmailDomain)
		} else {
			log.Println("定时任务 - 执行任务调度 - 新增用户操作")
			CreateUserProcess(PanelWebAddress, PanelWebPort, PanelWebUsername, PanelWebPassword, UserEmailDomain,
				VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName)
			log.Println("定时任务 - 执行任务调度 - 删除用户操作")
			DeleteUserProcess(PanelWebAddress, PanelWebPort, PanelWebUsername, PanelWebPassword,
				VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName,
				UserEmailDomain)
		}
		os.Remove(processTagFolder + "/TaskDispatcher")
	} else {
		log.Println("定时任务 - 不执行任务调度，因为有正在执行的任务", status)
	}
}
