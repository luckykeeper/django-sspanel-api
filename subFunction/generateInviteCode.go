// subFunction of vpnHelper
// Powered By Luckykeeper <luckykeeper@luckykeeper.site | https://luckykeeper.site>
package subFunction

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
	"vpnHelper/model"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
	_ "github.com/lib/pq"
)

var (
	GetInviteCodePage, InviteCodeLine             int
	GetInviteCodePageLastMark, InviteCodeLastMark bool
)

// 爬虫获取系统已有的邀请码
func GetInviteCode(PanelWebAddress, PanelWebPort, PanelWebUsername, PanelWebPassword,
	VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName string) {

	initializeTableInviteCode(VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName)

	// create context
	options := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true)) // debug(false)|prod(true)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), options...)
	defer cancel()
	ctx, _ := chromedp.NewContext(
		allocCtx,
	)
	// 添加超时时间
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var htmlContent []string
	// var htmlContent string

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

	listenPageStatusInviteCode(ctx, PanelWebAddress, PanelWebPort)

	// 管理面板 - 获取已有邀请码
	for GetInviteCodePage = 1; GetInviteCodePage >= 1; GetInviteCodePage++ {
		if err := chromedp.Run(ctx,
			chromedp.Navigate("http://"+PanelWebAddress+":"+PanelWebPort+"/admin/sspanel/invitecode/?p="+strconv.Itoa(GetInviteCodePage)+"&o=3.-2"),
			chromedp.WaitVisible(`#changelist-form > div.actions > button.el-button.el-button--primary.el-button--small`, chromedp.ByQuery),
			chromedp.Sleep(2*time.Second),
		); err != nil {
			log.Println(err)
		}
		var hContent string
		if !GetInviteCodePageLastMark {
			chromedp.Run(ctx,
				chromedp.OuterHTML("#changelist-form", &hContent, chromedp.ByQuery),
			)
			// htmlContent = htmlContent + hContent
			htmlContent = append(htmlContent, hContent)

		}
	}
	// 调试 -> 输出邀请码到文件（可以正确渲染）
	// fmt.Println(htmlContent)
	// file, _ := os.Create("./data/temp/htmlContent.html")
	// defer file.Close()
	// file.WriteString(htmlContent)

	// fmt.Println("htmlContentNum:", len(htmlContent))
	// fmt.Println("htmlContent1", htmlContent[7])

	var inviteCodeSum string

	for _, content := range htmlContent {
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
		if err != nil {
			log.Println(err)
		}
		// 多次循环重置状态
		InviteCodeLine = 1
		InviteCodeLastMark = false

		// 批量爬取邀请码
		for InviteCodeLine = 1; InviteCodeLine >= 1; InviteCodeLine++ {
			var inviteCode model.InviteCode
			inviteCode.InviteCode = doc.Find("#result_list > tbody > tr:nth-child(" + strconv.Itoa(InviteCodeLine) + ") > th > a").Text()
			inviteCode.CreateTime = doc.Find("#result_list > tbody > tr:nth-child(" + strconv.Itoa(InviteCodeLine) + ") > td.field-created_at.nowrap").Text()
			// goquery 定位参考 jquery 即可
			// https://blog.csdn.net/github_26672553/article/details/50747543
			inviteCode.CodeUsed, _ = doc.Find("#result_list > tbody > tr:nth-child(" + strconv.Itoa(InviteCodeLine) + ") > td.field-used > img").Attr("alt")
			inviteCode.CodeType = doc.Find("#result_list > tbody > tr:nth-child(" + strconv.Itoa(InviteCodeLine) + ") > td.field-code_type").Text()

			if inviteCode.InviteCode == "" {
				InviteCodeLine = -1
				InviteCodeLastMark = true
			}
			if !InviteCodeLastMark && inviteCode.CodeUsed == "False" {
				// 调试用
				// fmt.Println("______________________________")
				// fmt.Println("邀请码:", inviteCode.InviteCode)
				// fmt.Println("创建时间:", inviteCode.CreateTime)
				// fmt.Println("是否使用:", inviteCode.CodeUsed)
				// fmt.Println("类型", inviteCode.CodeType)

				// 2022/12/17 修改前
				// insertSql := "INSERT INTO invitecode (invitecode) values ('" + inviteCode.InviteCode + "');"
				// db, _ := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
				// 	VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName))
				// defer db.Close()
				// db.Exec(insertSql)

				// 将多个 SQL 语句合并为一句，减轻对服务器的负载
				inviteCodeSum = inviteCodeSum + "('" + inviteCode.InviteCode + "'),"
			}
		}
	}
	if len(inviteCodeSum) > 2 {
		// fmt.Println(inviteCodeSum)
		insertSql := "INSERT INTO invitecode (invitecode) values " + inviteCodeSum[0:len(inviteCodeSum)-1] + ";"
		db, _ := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName))
		defer db.Close()
		db.Exec(insertSql)
		log.Println("邀请码获取完成")
	} else {
		log.Println("本次邀请码获取出错，等待重试中")
		time.Sleep(60 * time.Second)
		log.Println("本次邀请码获取出错，等待重试中……等待时间结束，重试执行中")
		GetInviteCode(PanelWebAddress, PanelWebPort, PanelWebUsername, PanelWebPassword,
			VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName)
	}
}

// 监听状态，寻找最后页面
func listenPageStatusInviteCode(ctx context.Context, PanelWebAddress, PanelWebPort string) {
	// 判断端口号是否为 80 ，为 80 时不能加端口号否则无法监听到指定方法
	var eventRequestWillBeSentListenUrl string
	if PanelWebPort == "80" {
		eventRequestWillBeSentListenUrl = "http://" + PanelWebAddress + "/admin/sspanel/invitecode/?e=1"
	} else {
		eventRequestWillBeSentListenUrl = "http://" + PanelWebAddress + ":" + PanelWebPort + "/admin/sspanel/invitecode/?e=1"
	}
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {

		case *network.EventRequestWillBeSent:
			// 监听指定方法，在最后一页后给程序发送停止信号
			if ev.Request.URL == eventRequestWillBeSentListenUrl {
				GetInviteCodePage = -1           // 停止循环
				GetInviteCodePageLastMark = true // 停止输出
			}
		}
	})
}

// 获取数据库内邀请码个数
func GetInviteCodeNum(PanelWebAddress, PanelWebPort, PanelWebUsername, PanelWebPassword,
	VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName string) (inviteCodeNum int) {
	GetInviteCode(PanelWebAddress, PanelWebPort, PanelWebUsername, PanelWebPassword,
		VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName)
	db, _ := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName))
	defer db.Close()
	querySql := "SELECT COUNT(invitecode) FROM invitecode;"
	db.QueryRow(querySql).Scan(&inviteCodeNum)
	return
}

// 生成邀请码
func GenerateInviteCode(PanelWebAddress, PanelWebPort, PanelWebUsername, PanelWebPassword string, generateInviteCodeNum int) {

	// create context
	options := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true)) // debug(false)|prod(true)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), options...)
	defer cancel()
	ctx, _ := chromedp.NewContext(
		allocCtx,
	)
	// 添加超时时间
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
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
		chromedp.Navigate("http://"+PanelWebAddress+":"+PanelWebPort+"/my_admin/invite/"),
		chromedp.WaitVisible(`body > div > div > div.column.is-10 > div:nth-child(2) > div:nth-child(1) > form > div > div:nth-child(2) > div > select`),
		chromedp.SendKeys(`body > div > div > div.column.is-10 > div:nth-child(2) > div:nth-child(1) > form > div > div:nth-child(2) > div > select`, kb.ArrowDown, chromedp.ByQuery),
		chromedp.SendKeys(`body > div > div > div.column.is-10 > div:nth-child(2) > div:nth-child(1) > form > div > div:nth-child(1) > input`, strconv.Itoa(generateInviteCodeNum), chromedp.ByQuery),
		chromedp.Click(`body > div > div > div.column.is-10 > div:nth-child(2) > div:nth-child(1) > form > div > div:nth-child(3) > button`, chromedp.ByQuery),
		chromedp.WaitVisible(`body > div.swal-overlay.swal-overlay--show-modal > div > div.swal-footer > div > button`, chromedp.ByQuery),
	); err != nil {
		log.Println(err)
	}
	log.Println("添加邀请码完成")
}
