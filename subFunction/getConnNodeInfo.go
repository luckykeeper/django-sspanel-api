// subFunction of vpnHelper
// Powered By Luckykeeper <luckykeeper@luckykeeper.site | https://luckykeeper.site>
package subFunction

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
	"vpnHelper/model"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	_ "github.com/lib/pq"
)

var (
	GetNodeInfoPage         int
	GetNodeInfoPageLastMark bool
)

// 爬取代理节点信息
func GetConnNodeInfo(PanelWebAddress, PanelWebPort, PanelWebUsername, PanelWebPassword,
	VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName string) {

	if getConnNodeInfoProcessing, _ := PathExists(processTagFolder + "/getConnNodeInfoProcessing"); !getConnNodeInfoProcessing {
		os.Create(processTagFolder + "/getConnNodeInfoProcessing")
	}

	// 记录运行时间
	getConnNodeInfoTagFile, _ := os.OpenFile(processTagFolder+"/getConnNodeInfoTag", os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0777)
	defer getConnNodeInfoTagFile.Close()

	// 先写一个时间，防止反复判定运行(重试时间：120分钟)
	fakeTimeStr := time.Now().Add(-22 * time.Hour).Format(time.RFC822)
	getConnNodeInfoTagFile.WriteString(fakeTimeStr)

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

	listenPageStatusNodeInfo(ctx, PanelWebAddress, PanelWebPort)

	// 管理面板 - 获取节点信息 （注意仅适用于 SS 节点，使用其它类型节点需要适当改写此方法）
	for GetNodeInfoPage = 1; GetNodeInfoPage >= 1; GetNodeInfoPage++ {
		if err := chromedp.Run(ctx,
			// 跳转到节点信息页
			chromedp.Navigate("http://"+PanelWebAddress+":"+PanelWebPort+"/admin/proxy/proxynode/"+strconv.Itoa(GetNodeInfoPage)+"/change/"),
			chromedp.Sleep(2*time.Second),
		); err != nil {
			log.Println(err)
		}
		var hContent string

		if !GetNodeInfoPageLastMark {
			chromedp.Run(ctx,
				chromedp.OuterHTML("#proxynode_form > div", &hContent, chromedp.ByQuery),
			)

			htmlContent = append(htmlContent, hContent)

		}
	}

	for _, content := range htmlContent {
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
		if err != nil {
			log.Println(err)
		}

		var (
			nodeInfo, nodeSqlInfo model.NodeInfo
			nodeInfos             string
		)
		nodeInfo.Sequence, _ = doc.Find("#id_sequence").Attr("value")
		nodeInfo.Name, _ = doc.Find("#id_name").Attr("value")
		nodeInfo.Ip, _ = doc.Find("#id_server").Attr("value")
		// jQuery 取选项已选值参考：https://qastack.cn/programming/13089944/jquery-get-selected-option-value-not-the-text-but-the-attribute-value
		// https://blog.csdn.net/syq8023/article/details/93083172 <- 概述
		// https://zhuanlan.zhihu.com/p/98912885 <- 此处用这个
		nodeInfo.NodeType = doc.Find("#id_node_type option:checked").Text()
		nodeInfo.Tips, _ = doc.Find("#id_info").Attr("value")
		nodeInfo.Method = doc.Find("#id_ss_config-0-method option:checked").Text()
		nodeInfo.Port, _ = doc.Find("#id_ss_config-0-multi_user_port").Attr("value")

		nodeInfos = nodeInfos + "('" + nodeInfo.Sequence + "'," +
			"'" + nodeInfo.Name + "'," +
			"'" + nodeInfo.Ip + "'," +
			"'" + nodeInfo.NodeType + "'," +
			"'" + nodeInfo.Tips + "'," +
			"'" + nodeInfo.Method + "'," +
			"'" + nodeInfo.Port +
			"')"
		db, _ := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName))
		defer db.Close()

		querySql := "SELECT sequence,name,ip,nodetype,tips,method,port FROM nodeinfo WHERE name='" + nodeInfo.Name + "';"
		queryResult := db.QueryRow(querySql).Scan(&nodeSqlInfo.Sequence, &nodeSqlInfo.Name, &nodeSqlInfo.Ip, &nodeSqlInfo.NodeType, &nodeSqlInfo.Tips, &nodeSqlInfo.Method, &nodeSqlInfo.Port)
		if queryResult == sql.ErrNoRows {
			insertSql := "INSERT INTO nodeinfo (sequence,name,ip,nodetype,tips,method,port) values " + nodeInfos + ";"
			db.Exec(insertSql)
			fmt.Println("数据库新增节点记录：", nodeInfo.Name)

			// 注意！如果删除了已经被爬取的节点，必须到数据库手动删除数据，否则会导致用户获得过时的连接信息
		} else {
			updateSql := "UPDATE nodeinfo SET sequence='" + nodeInfo.Sequence + "'," +
				"name='" + nodeInfo.Name + "'," +
				"ip='" + nodeInfo.Ip + "'," +
				"nodetype='" + nodeInfo.NodeType + "'," +
				"tips='" + nodeInfo.Tips + "'," +
				"method='" + nodeInfo.Method + "'," +
				"port='" + nodeInfo.Port + "';"
			db.Exec(updateSql)
			log.Println("数据库更新节点记录：", nodeInfo.Name)
		}
	}
	getConnNodeInfoTagFile1, _ := os.OpenFile(processTagFolder+"/getConnNodeInfoTag", os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0777)
	defer getConnNodeInfoTagFile1.Close()
	doneTime := time.Now().Format(time.RFC822)
	getConnNodeInfoTagFile1.WriteString(doneTime)

	os.Remove(processTagFolder + "/getConnNodeInfoProcessing")
	log.Println("节点信息获取完成")
}

// 监听状态，寻找最后页面
func listenPageStatusNodeInfo(ctx context.Context, PanelWebAddress, PanelWebPort string) {
	// 判断端口号是否为 80 ，为 80 时不能加端口号否则无法监听到指定方法
	var eventRequestWillBeSentListenUrl string
	if PanelWebPort == "80" {
		eventRequestWillBeSentListenUrl = "http://" + PanelWebAddress + "/admin/"
	} else {
		eventRequestWillBeSentListenUrl = "http://" + PanelWebAddress + ":" + PanelWebPort + "/admin/"
	}
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {

		case *network.EventRequestWillBeSent:
			// 监听指定方法，在最后一页后给程序发送停止信号
			if ev.Request.URL == eventRequestWillBeSentListenUrl {
				GetNodeInfoPage = -1           // 停止循环
				GetNodeInfoPageLastMark = true // 停止输出
			}
		}
	})
}
