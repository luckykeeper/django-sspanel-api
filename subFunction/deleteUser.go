// subFunction of vpnHelper
// Powered By Luckykeeper <luckykeeper@luckykeeper.site | https://luckykeeper.site>
package subFunction

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
)

// 爬虫删除用户 - 流程入口
func DeleteUserProcess(PanelWebAddress, PanelWebPort, PanelWebUsername, PanelWebPassword,
	VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName,
	UserEmailDomain string) {
	// 入口都应该在 Tag 内添加时间戳，5分钟CD(deleteUser/createUser)
	// 尝试读取、判断是否近期（正在）运行，防止反复运行
	var runDurationDeleteUserProcess time.Duration
	if deleteUserProcessing, _ := PathExists(processTagFolder + "/deleteUserProcessingTag"); deleteUserProcessing {
		lastTimeRunFile, _ := os.ReadFile(processTagFolder + "/deleteUserProcessingTag")

		lastTimeRunStr := string(lastTimeRunFile)
		lastTimeRun, _ := time.Parse(time.RFC822, lastTimeRunStr)

		runDurationDeleteUserProcess = time.Since(lastTimeRun)

	} else {
		// 没有 Tag 直接运行
		runDurationDeleteUserProcess = runDurationDeleteUserProcess + 6*time.Minute
	}
	if runDurationDeleteUserProcess > 5*time.Minute {
		// 记录运行时间
		deleteUserProcessTagFile, _ := os.OpenFile(processTagFolder+"/deleteUserProcessingTag", os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0777)
		defer deleteUserProcessTagFile.Close()
		// 写入程序开始运行时间
		runTimeStr := time.Now().Format(time.RFC822)
		deleteUserProcessTagFile.WriteString(runTimeStr)

		log.Println("删除流程正式开始")

		queryIfExistUserToBeDeleteSql := "SELECT COUNT(username) FROM vpnhelper WHERE status='markRemove';"
		db, _ := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName))
		defer db.Close()
		var queryIfExistUserToBeDeleteResult int
		db.QueryRow(queryIfExistUserToBeDeleteSql).Scan(&queryIfExistUserToBeDeleteResult)
		if queryIfExistUserToBeDeleteResult == 0 {
			if deleteUserProcessing, _ := PathExists(processTagFolder + "/deleteUserProcessing"); deleteUserProcessing {
				log.Println("删除过期的删除流程 Tag 标记")
				os.Remove(processTagFolder + "/deleteUserProcess")
			} else {
				log.Println("不存在将被删除的用户，略过删除流程")
			}
		} else {
			deleteUser(PanelWebAddress, PanelWebPort, PanelWebUsername, PanelWebPassword,
				VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName,
				UserEmailDomain)
		}
		os.Remove(processTagFolder + "/deleteUserProcess")
	}
}

// 爬虫删除用户
func deleteUser(PanelWebAddress, PanelWebPort, PanelWebUsername, PanelWebPassword,
	VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName,
	UserEmailDomain string) {

	queryExistsUserToBeDeleteSql := "SELECT COUNT(username) FROM vpnhelper WHERE status='markRemove';"
	var queryExistsUserToBeDeleteResult int
	db, _ := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName))
	defer db.Close()
	db.QueryRow(queryExistsUserToBeDeleteSql).Scan(&queryExistsUserToBeDeleteResult)
	if queryExistsUserToBeDeleteResult != 0 {

		// create context
		options := append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", true)) // debug(false)|prod(true)
		allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), options...)
		defer cancel()
		ctx, _ := chromedp.NewContext(
			allocCtx,
		)
		// 添加超时时间
		ctx, cancel = context.WithTimeout(ctx, 50*time.Second)
		defer cancel()

		if err := chromedp.Run(ctx,
			// 管理面板 - 登录页配置
			chromedp.EmulateViewport(1920, 1080),
			chromedp.Navigate("http://"+PanelWebAddress+":"+PanelWebPort+"/login/"),
			chromedp.WaitVisible(`#id_username`),
			chromedp.SendKeys(`#id_username`, PanelWebUsername, chromedp.ByQuery),
			chromedp.SendKeys(`#id_password`, PanelWebPassword, chromedp.ByQuery),
			chromedp.Click(`body > div > div > div.column.is-10 > div:nth-child(2) > div > form > div > p:nth-child(1) > button`, chromedp.ByQuery),
			chromedp.WaitVisible(`body > div.swal-overlay.swal-overlay--show-modal > div > div.swal-footer > div > button`),
			chromedp.Click(`body > div.swal-overlay.swal-overlay--show-modal > div > div.swal-footer > div > button`, chromedp.ByQuery),
		); err != nil {
			log.Println(err)
		}

		// 限制执行数量，防止 Chrome 占用大量内存
		// queryUserToBeDeleteSql := "SELECT username,requesttime FROM vpnhelper WHERE status='markRemove';"
		queryUserToBeDeleteSql := "SELECT username,requesttime FROM vpnhelper WHERE status='markRemove' ORDER BY requesttime LIMIT 5;"
		db, _ := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName))
		defer db.Close()
		rows, _ := db.Query(queryUserToBeDeleteSql)
		defer rows.Close()

		// 管理面板 - 搜索要删除的用户
		// 注意当天创建的用户当天不能删除（面板报 500）
		for rows.Next() {
			var username, htmlContent, requestTimeStr string
			rows.Scan(&username, &requestTimeStr)
			requestTime, _ := time.Parse(time.RFC822, requestTimeStr)
			runDuration := time.Since(requestTime)
			if runDuration > 48*time.Hour {
				log.Println("用户：" + username + "进入删除流程!")
				if err := chromedp.Run(ctx,
					// 用邮箱搜索用户，能得到唯一值（yuuka：乐）
					chromedp.Navigate("http://"+PanelWebAddress+":"+PanelWebPort+"/admin/sspanel/user/?q="+username+"%40"+UserEmailDomain),
					chromedp.Sleep(2*time.Second),
					chromedp.OuterHTML("#changelist-form", &htmlContent, chromedp.ByQuery),
				); err != nil {
					log.Println(err)
				}

				doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
				if err != nil {
					log.Println(err)
				}
				searchUser := doc.Find("#result_list > tbody > tr > th > a").Text()
				var hContent string
				if searchUser == username {
					if err := chromedp.Run(ctx,
						chromedp.Navigate("http://"+PanelWebAddress+":"+PanelWebPort+"/admin/sspanel/user/?q="+username+"%40"+UserEmailDomain),
						chromedp.Sleep(1*time.Second),
						chromedp.Click(`#result_list > tbody > tr > td.action-checkbox > input`, chromedp.ByQuery),
						chromedp.Click(`#changelist-form > div.actions > button.el-button.stop-submit.el-button--danger.el-button--small`, chromedp.ByQuery),
						chromedp.Click(`body > div.el-message-box__wrapper > div > div.el-message-box__btns > button.el-button.el-button--default.el-button--small.el-button--primary`, chromedp.ByQuery),
						chromedp.Click(`#content > form > div > input[type=submit]:nth-child(4)`, chromedp.ByQuery),
						chromedp.Sleep(1*time.Second),
						chromedp.OuterHTML("body > div.el-notification.right", &hContent, chromedp.ByQuery),
					); err != nil {
						log.Println(err)
					}
					if hContent != "" {
						deleteSql := "DELETE FROM vpnhelper WHERE username='" + username + "';"
						db, _ := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
							VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName))
						defer db.Close()
						db.Exec(deleteSql)
						log.Println("用户：" + username + "已删除成功!")
					} else {
						updateSql := "UPDATE vpnhelper SET statuscode=500 WHERE username='" + username + "';"
						db.Exec(updateSql)
						log.Println("用户：" + username + "删除失败，可能是系统数据未刷新阻止了删除")
					}
				}
			}
		}
	}
}
