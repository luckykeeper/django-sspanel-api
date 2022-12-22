// subFunction of vpnHelper
// Powered By Luckykeeper <luckykeeper@luckykeeper.site | https://luckykeeper.site>
package subFunction

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	_ "github.com/lib/pq"
)

// 用户注册流程总入口
func CreateUserProcess(PanelWebAddress, PanelWebPort, PanelWebUsername, PanelWebPassword, UserEmailDomain,
	VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName string) {
	// 入口都应该在Tag内添加时间戳，5分钟CD(deleteUser/createUser)
	// 尝试读取、判断是否近期（正在）运行，防止反复运行
	var runDurationCreateUserProcess time.Duration
	if createUserProcessing, _ := PathExists(processTagFolder + "/CreateUserProcessTag"); createUserProcessing {
		lastTimeRunFile, _ := os.ReadFile(processTagFolder + "/CreateUserProcessTag")

		lastTimeRunStr := string(lastTimeRunFile)
		lastTimeRun, _ := time.Parse(time.RFC822, lastTimeRunStr)

		runDurationCreateUserProcess = time.Since(lastTimeRun)

	} else {
		// 没有 Tag 直接运行
		runDurationCreateUserProcess = runDurationCreateUserProcess + 6*time.Minute
	}
	if runDurationCreateUserProcess > 5*time.Minute {
		// 记录运行时间
		createUserProcessTagFile, _ := os.OpenFile(processTagFolder+"/CreateUserProcessTag", os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0777)
		defer createUserProcessTagFile.Close()
		// 写入程序开始运行时间
		runTimeStr := time.Now().Format(time.RFC822)
		createUserProcessTagFile.WriteString(runTimeStr)

		if userToBeAddCount := MarkUserToBeAdd(VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName); userToBeAddCount != 0 {
			if createUserProcessing, _ := PathExists(processTagFolder + "/CreateUserProcess"); !createUserProcessing {
				os.Create(processTagFolder + "/CreateUserProcess")
			}
			log.Println("开始创建用户流程")
			for i := 1; i >= 1; i++ {
				// userToBeAddCount := MarkUserToBeAdd(VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName)
				nowInviteCodeNum := GetInviteCodeNum(PanelWebAddress, PanelWebPort, PanelWebUsername, PanelWebPassword,
					VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName)
				if userToBeAddCount > nowInviteCodeNum {
					log.Println("当前邀请码数量", nowInviteCodeNum, "。当前要新增的用户数量", userToBeAddCount)
					newInviteCodeNum := userToBeAddCount - nowInviteCodeNum
					log.Println("当前邀请码数量不足，正在添加", newInviteCodeNum, "个邀请码")
					GenerateInviteCode(PanelWebAddress, PanelWebPort, PanelWebUsername, PanelWebPassword, newInviteCodeNum)
				} else {
					// 邀请码数量足够，进入下一流程
					i = -1
				}
			}
			log.Println("当前邀请码数量足够，开始新增用户流程")
			queryUserNameSql := "SELECT username FROM vpnHelper WHERE status='processing' AND statuscode=102;"
			db, _ := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
				VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName))
			defer db.Close()
			rows, _ := db.Query(queryUserNameSql)
			defer rows.Close()
			for rows.Next() {
				var userName string
				rows.Scan(&userName)
				log.Println("将添加用户：", userName)
				statusCode, loginPassword, connPassword := CreateUser(PanelWebAddress, PanelWebPort, userName, UserEmailDomain,
					VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName)
				if statusCode == 200 {
					db, _ := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
						VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName))
					defer db.Close()
					db.Exec("UPDATE vpnhelper SET status='exists',statuscode=200,loginpasswd='" + loginPassword + "'," + "connpasswd='" + connPassword + "'" + " WHERE userName='" + userName + "';")
					log.Println("用户：", userName, "已成功添加")
				} else if statusCode == 409 {
					db, _ := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
						VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName))
					defer db.Close()
					// 标记状态码 202 ，稍后可进行删除重试处理（由 OA 侧发起）
					db.Exec("UPDATE vpnhelper SET status='exists',statuscode=202 WHERE userName='" + userName + "';")
					log.Println("用户：", userName, "存在,已在数据库中标记为完成")
				} else if statusCode == 500 {
					logfile, _ := os.OpenFile("./data/dataNeedNotice/CreateUserFailed.log", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0777)
					logfile.WriteString(userName + ",")
					log.Println("用户：", userName, "添加失败,服务器内部错误")
				}
			}
			os.Remove(processTagFolder + "/CreateUserProcess")
		} else {
			log.Println("系统内无新增用户需求，略过新增用户流程")
			// if createUserProcessing, _ := PathExists(processTagFolder + "/CreateUserProcess"); createUserProcessing {
			os.Remove(processTagFolder + "/CreateUserProcess")
			// log.Println("已移除过期的新增用户流程进行 Tag")
			// }
		}
	}
}

// 用户注册
func CreateUser(PanelWebAddress, PanelWebPort, userName, UserEmailDomain,
	VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName string) (status int, loginPassword, connPassword string) {
	// create context
	options := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true)) // debug(false)|prod(true)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), options...)
	defer cancel()
	ctx, _ := chromedp.NewContext(
		allocCtx,
	)
	// 添加超时时间
	ctx, cancel = context.WithTimeout(ctx, 240*time.Second)
	defer cancel()
	newUserPasswd := GenerateUserPassword()
	var hContent string
	var inviteCode string
	db, _ := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName))
	defer db.Close()
	db.QueryRow("SELECT inviteCode FROM inviteCode LIMIT 1;").Scan(&inviteCode)
	if err := chromedp.Run(ctx,
		// 管理面板 - 注册页配置
		chromedp.EmulateViewport(1920, 1080),
		chromedp.Navigate("http://"+PanelWebAddress+":"+PanelWebPort+"/register/"),
		chromedp.WaitVisible(`#id_invitecode`),
		chromedp.SendKeys(`#id_username`, userName, chromedp.ByQuery),
		chromedp.SendKeys(`#id_email`, userName+"@"+UserEmailDomain, chromedp.ByQuery),
		chromedp.SendKeys(`#id_password1`, newUserPasswd, chromedp.ByQuery),
		chromedp.SendKeys(`#id_password2`, newUserPasswd, chromedp.ByQuery),
		chromedp.SendKeys(`#id_invitecode`, inviteCode, chromedp.ByQuery),
		chromedp.Click(`body > div > div > div.column.is-10 > div:nth-child(2) > div > form > div > p:nth-child(1) > button`, chromedp.ByQuery),
		chromedp.Sleep(time.Second*2),
		chromedp.Navigate("http://"+PanelWebAddress+":"+PanelWebPort),
		chromedp.OuterHTML(`body`, &hContent, chromedp.ByQuery),
		// chromedp.WaitVisible(`body > div.swal-overlay.swal-overlay--show-modal > div > div.swal-footer > div > button`),
	); err != nil {
		log.Println(err)
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(hContent))
	if err != nil {
		log.Println(err)
	}
	if doc.Find("body > div > div > div.column.is-10 > div.box > button > a").Text() == "进入" {
		// 成功
		db, _ := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName))
		defer db.Close()
		db.Exec("DELETE FROM inviteCode WHERE inviteCode='" + inviteCode + "';")
		log.Println("用户 " + userName + " 注册成功！")

		// 获取连接密码
		var outerConnPasswd string
		chromedp.Run(ctx,
			chromedp.Navigate("http://"+PanelWebAddress+":"+PanelWebPort+"/users/userinfo/"),
			chromedp.WaitVisible(`body > div.container > div > div.column.is-10 > div:nth-child(2) > div > div.tile.is-7.is-vertical.is-parent > div > article.message.is-success > div.message-body > li:nth-child(2) > code`),
			chromedp.OuterHTML(`body > div.container > div > div.column.is-10 > div:nth-child(2) > div > div.tile.is-7.is-vertical.is-parent > div > article.message.is-success > div.message-body > li:nth-child(2) > code`, &outerConnPasswd, chromedp.ByQuery),
		)
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(outerConnPasswd))
		if err != nil {
			log.Println(err)
		}
		connPassword = doc.Text()
		log.Println("用户 " + userName + " 的连接密码获取成功！")

		// 注册完成需要退出
		chromedp.Run(ctx,
			chromedp.Navigate("http://"+PanelWebAddress+":"+PanelWebPort+"/logout/"),
		)
		return 200, newUserPasswd, connPassword
	} else if doc.Find("body > div.container > div > div.column.is-10 > div:nth-child(3) > a").Text() == "注册" {
		// 失败
		log.Println("用户 " + userName + " 已经存在,不能重复注册")
		return 409, "", ""
	} else {
		return 500, "", ""
	}
}

// 生成随机密码
func GenerateUserPassword() (userPasswd string) {
	rand.Seed(time.Now().UnixNano())
	uLen := 32
	b := make([]byte, uLen)
	rand.Read(b)
	rand_str := hex.EncodeToString(b)[0:uLen]
	return rand_str
}

// 标记待添加用户
func MarkUserToBeAdd(VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername,
	VpnHelperDBPassword, VpnHelperDBName string) (userCount int) {
	// 标记表内已提交用户，开始处理（processing-102）
	log.Println("开始标记待添加用户进入添加流程")
	db, _ := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName))
	defer db.Close()
	markSql := "UPDATE vpnhelper set status='processing',statuscode=102 WHERE status='submit';"
	db.Exec(markSql)
	log.Println("标记待添加用户进入添加流程成功完成")
	var countProcessingUser int
	countProcessingUserSql := "SELECT count(username) FROM vpnHelper WHERE status='processing' AND statuscode=102;"
	db.QueryRow(countProcessingUserSql).Scan(&countProcessingUser)
	log.Println("当前进入添加流程用户数量：", countProcessingUser)
	return countProcessingUser
}
