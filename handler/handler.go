// Handler of vpnHelper
// Powered By Luckykeeper <luckykeeper@luckykeeper.site | https://luckykeeper.site>
package handler

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
	"vpnHelper/model"
	subFunction "vpnHelper/subFunction"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"gopkg.in/ini.v1"
)

var VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName,
	PanelWebUsername string

func init() {
	configFile, err := ini.Load("./config.ini")
	if err != nil {
		fmt.Printf("读取配置文件 config.ini 失败，失败原因为: %v", err)
		os.Exit(1)
	}
	PanelWebUsername = configFile.Section("panelWeb").Key("username").String()

	VpnHelperDBAddress = configFile.Section("vpnHelper").Key("dbAddress").String()
	VpnHelperDBPort = configFile.Section("vpnHelper").Key("dbPort").String()
	VpnHelperDBUsername = configFile.Section("vpnHelper").Key("dbUsername").String()
	VpnHelperDBPassword = configFile.Section("vpnHelper").Key("dbPassword").String()
	VpnHelperDBName = configFile.Section("vpnHelper").Key("dbName").String()
}

// Handler - 生成用户
// 传值：Token、用户名
func GenerateUserProfile(context *gin.Context) {
	var user model.UserModel
	context.ShouldBind(&user)

	var result model.ResultMsg
	// Token 错误
	if !subFunction.CheckToken(user.Token,
		VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName) {
		result.StatusCode = 401
		result.StatusString = "Unauthorized Token"
		context.JSON(http.StatusOK, result)
	} else if user.UserName == PanelWebUsername { // 面板的管理员账户不允许操作
		result.StatusCode = 406
		result.StatusString = "Not Acceptable"
		context.JSON(http.StatusOK, result)
	} else {
		// Token 正确，用户可操作，继续执行
		// 先查询该用户是否存在系统内
		db, _ := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName))
		defer db.Close()
		queryUserIsInDatabaseSql := "SELECT COUNT(username) FROM vpnhelper WHERE username='" + user.UserName + "';"
		queryUserInDatabaseStatusSql := "SELECT status,statuscode FROM vpnhelper WHERE username='" + user.UserName + "';"
		var queryUserIsInDatabaseResult int
		var (
			queryUserInDatabaseStatus     string
			queryUserInDatabaseStatusCode int
		)
		db.QueryRow(queryUserIsInDatabaseSql).Scan(&queryUserIsInDatabaseResult)
		db.QueryRow(queryUserInDatabaseStatusSql).Scan(&queryUserInDatabaseStatus, &queryUserInDatabaseStatusCode)
		if queryUserIsInDatabaseResult == 0 {
			insertSql := "INSERT INTO vpnHelper (username, status) values ('" + user.UserName + "','submit');"
			db, _ := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
				VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName))
			defer db.Close()
			_, err := db.Exec(insertSql)
			if err == nil {
				result.StatusCode = 200
				result.StatusString = "提交新增用户任务成功！"
				context.JSON(http.StatusOK, result)
				log.Println("从对端收到新增：" + user.UserName + "的请求，已处理")
			} else {
				result.StatusCode = 429
				result.StatusString = "服务请求过于频繁，请稍后再试"
				context.JSON(http.StatusOK, result)
				log.Println("从对端收到新增："+user.UserName+"的请求，但是插入数据时发生错误：", err)
			}
		} else if queryUserInDatabaseStatus == "submit" {
			result.StatusCode = 201
			result.StatusString = "该新增用户数据已提交至系统，正在排队处理中，请等待"
			context.JSON(http.StatusOK, result)
		} else if queryUserInDatabaseStatus == "processing" {
			result.StatusCode = 102
			result.StatusString = "该新增用户数据已提交至系统，并且正在处理中，请等待"
			context.JSON(http.StatusOK, result)
		} else if queryUserInDatabaseStatus == "exists" && queryUserInDatabaseStatusCode == 200 {
			result.StatusCode = 405
			result.StatusString = "用户数据已经存在于系统内，不允许重复提交！"
			context.JSON(http.StatusOK, result)
		} else if queryUserInDatabaseStatus == "exists" && queryUserInDatabaseStatusCode == 202 {
			result.StatusCode = 409
			result.StatusString = "用户数据已经存在于系统内，但是存在问题，请尝试发起删除用户流程后重新创建（重试）"
			context.JSON(http.StatusOK, result)
		} else if queryUserInDatabaseStatus == "markRemove" && queryUserInDatabaseStatusCode == 200 {
			result.StatusCode = 418
			result.StatusString = "用户数据正在删除,请等待删除操作完成。删除操作需要等待四十八小时后才会执行，请耐心等待"
			context.JSON(http.StatusOK, result)
		} else if queryUserInDatabaseStatus == "markRemove" && queryUserInDatabaseStatusCode == 500 {
			result.StatusCode = 500
			result.StatusString = "用户数据删除请求已提交满四十八小时，但是API尝试删除用户失败，这是由于系统内尚未刷新该数据允许删除，API会自动重试直到系统允许删除，请等待"
			context.JSON(http.StatusOK, result)
		}
	}
}

// Handler - 获取用户连接信息
// 传值：Token、用户名
func GetConnInfoByUser(context *gin.Context) {
	var user model.UserModel
	context.ShouldBind(&user)

	var result model.ResultMsg
	// Token 错误
	if !subFunction.CheckToken(user.Token,
		VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName) {
		result.StatusCode = 401
		result.StatusString = "Unauthorized Token"
		context.JSON(http.StatusOK, result)
	} else if user.UserName == PanelWebUsername { // 面板的管理员账户不允许操作
		result.StatusCode = 406
		result.StatusString = "Not Acceptable"
		context.JSON(http.StatusOK, result)
	} else {
		// Token 正确，用户可操作，继续执行
		// 先查询系统内是否存在节点
		db, _ := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName))
		defer db.Close()
		queryNodeIsInDatabaseSql := "SELECT COUNT(name) FROM nodeinfo;"
		var queryNodeIsInDatabaseSqlResult int
		db.QueryRow(queryNodeIsInDatabaseSql).Scan(&queryNodeIsInDatabaseSqlResult)
		if queryNodeIsInDatabaseSqlResult == 0 {
			result.StatusCode = 412
			result.StatusString = "系统内没有节点或 API 还未拉取节点信息，请稍后再试"
			context.JSON(http.StatusOK, result)
		} else { // 验证用户
			queryUserInDatabaseStatusSql := "SELECT status,statuscode FROM vpnhelper WHERE username='" + user.UserName + "';"
			var (
				queryUserInDatabaseStatus     string
				queryUserInDatabaseStatusCode int
			)
			db, _ := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
				VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName))
			defer db.Close()
			db.QueryRow(queryUserInDatabaseStatusSql).Scan(&queryUserInDatabaseStatus, &queryUserInDatabaseStatusCode)
			if queryUserInDatabaseStatus == "" {
				result.StatusCode = 204
				result.StatusString = "该用户不存在，无法获取连接信息"
				context.JSON(http.StatusOK, result)
			} else if queryUserInDatabaseStatus == "submit" {
				result.StatusCode = 201
				result.StatusString = "该新增用户数据已提交至系统，正在排队处理中，请等待"
				context.JSON(http.StatusOK, result)
			} else if queryUserInDatabaseStatus == "processing" {
				result.StatusCode = 102
				result.StatusString = "该新增用户数据已提交至系统，并且正在处理中，请等待"
				context.JSON(http.StatusOK, result)
			} else if queryUserInDatabaseStatus == "exists" && queryUserInDatabaseStatusCode == 200 {
				var (
					connInfoByUser           model.NodeInfo
					connInfoByUserNameReturn model.ConnInfoByUserName
				)
				var connPasswd string
				queryNodeInfoSql := "SELECT sequence,name,ip,nodetype,tips,method,port FROM nodeinfo;"
				row, _ := db.Query(queryNodeInfoSql)
				for row.Next() {
					row.Scan(&connInfoByUser.Sequence, &connInfoByUser.Name, &connInfoByUser.Ip, &connInfoByUser.NodeType, &connInfoByUser.Tips, &connInfoByUser.Method, &connInfoByUser.Port)
					queryConnPasswordSql := "SELECT connpasswd FROM vpnhelper WHERE username='" + user.UserName + "';"
					db.QueryRow(queryConnPasswordSql).Scan(&connPasswd)
					// url 编码方法应当使用 url.PathEscape ，参考：https://segmentfault.com/a/1190000040919065?sort=newest
					connInfoByUser.ConnInfo = connInfoByUser.NodeType + "://" + base64.StdEncoding.EncodeToString([]byte(connInfoByUser.Method+":"+connPasswd+"@"+connInfoByUser.Ip+":"+connInfoByUser.Port)) + "#" + url.PathEscape(connInfoByUser.Name)
					connInfoByUserNameReturn.NodeInfos = append(connInfoByUserNameReturn.NodeInfos, connInfoByUser)
				}
				result.StatusCode = 200
				result.StatusString = "查询成功!"
				connInfoByUserNameReturn.ResultMsg = append(connInfoByUserNameReturn.ResultMsg, result)
				context.JSON(http.StatusOK, connInfoByUserNameReturn)
			} else if queryUserInDatabaseStatus == "exists" && queryUserInDatabaseStatusCode == 202 {
				result.StatusCode = 409
				result.StatusString = "用户数据已经存在于系统内，但是可能存在问题，请尝试发起删除用户流程后重新创建（重试）"
				context.JSON(http.StatusOK, result)
			} else if queryUserInDatabaseStatus == "markRemove" {
				result.StatusCode = 418
				result.StatusString = "用户数据正在删除,不允许获取连接信息"
				context.JSON(http.StatusOK, result)
			}
		}
	}
}

// Handler - 删除用户
// 传值：Token、用户名
func DelUserProfile(context *gin.Context) {
	var user model.UserModel
	context.ShouldBind(&user)

	var result model.ResultMsg
	// Token 错误
	if !subFunction.CheckToken(user.Token,
		VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName) {
		result.StatusCode = 401
		result.StatusString = "Unauthorized Token"
		context.JSON(http.StatusOK, result)
	} else if user.UserName == PanelWebUsername { // 面板的管理员账户不允许操作
		result.StatusCode = 406
		result.StatusString = "Not Acceptable"
		context.JSON(http.StatusOK, result)
	} else {
		// Token 正确，用户可操作，继续执行
		// 先查询该用户是否存在系统内
		db, _ := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName))
		defer db.Close()
		queryUserIsInDatabaseSql := "SELECT COUNT(username) FROM vpnhelper WHERE username='" + user.UserName + "';"
		queryUserInDatabaseStatusSql := "SELECT status,statuscode FROM vpnhelper WHERE username='" + user.UserName + "';"
		var queryUserIsInDatabaseResult int
		var (
			queryUserInDatabaseStatus     string
			queryUserInDatabaseStatusCode int
		)
		db.QueryRow(queryUserIsInDatabaseSql).Scan(&queryUserIsInDatabaseResult)
		db.QueryRow(queryUserInDatabaseStatusSql).Scan(&queryUserInDatabaseStatus, &queryUserInDatabaseStatusCode)
		if queryUserIsInDatabaseResult == 0 {
			result.StatusCode = 404
			result.StatusString = "系统内没有该用户记录!"
			context.JSON(http.StatusOK, result)

		} else if queryUserInDatabaseStatus == "submit" {
			result.StatusCode = 201
			result.StatusString = "该新增用户数据已提交至系统，正在排队处理中，请等待"
			context.JSON(http.StatusOK, result)
		} else if queryUserInDatabaseStatus == "processing" {
			result.StatusCode = 102
			result.StatusString = "该新增用户数据已提交至系统，并且正在处理中，请等待"
			context.JSON(http.StatusOK, result)
		} else if queryUserInDatabaseStatus == "exists" && queryUserInDatabaseStatusCode == 200 {
			requestTime := time.Now().Format(time.RFC822)
			markSql := "UPDATE vpnHelper SET status='markRemove',requesttime='" + requestTime + "' WHERE username='" + user.UserName + "';"
			db, _ := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
				VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName))
			defer db.Close()
			_, err := db.Exec(markSql)
			if err == nil {
				result.StatusCode = 200
				result.StatusString = "系统已收到删除用户：" + user.UserName + " 的请求，已处理"
				context.JSON(http.StatusOK, result)
				log.Println("从对端收到删除：" + user.UserName + "的请求，已处理")
			} else {
				result.StatusCode = 429
				result.StatusString = "服务请求过于频繁，请稍后再试"
				context.JSON(http.StatusOK, result)
				log.Println("从对端收到删除："+user.UserName+"的请求，但是标记数据时发生错误：", err)
			}
		} else if queryUserInDatabaseStatus == "exists" && queryUserInDatabaseStatusCode == 202 {
			requestTime := time.Now().Format(time.RFC822)
			markSql := "UPDATE vpnHelper SET status='markRemove',requesttime='" + requestTime + "' WHERE username='" + user.UserName + "';"
			db, _ := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
				VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName))
			defer db.Close()
			_, err := db.Exec(markSql)
			if err == nil {
				result.StatusCode = 200
				result.StatusString = "系统已收到删除可能存在问题的用户：" + user.UserName + " 的请求，已处理"
				context.JSON(http.StatusOK, result)
				log.Println("从对端收到删除可能存在问题的用户：" + user.UserName + "的请求，已处理")
			} else {
				result.StatusCode = 429
				result.StatusString = "服务请求过于频繁，请稍后再试"
				context.JSON(http.StatusOK, result)
				log.Println("从对端收到删除可能存在问题的用户："+user.UserName+"的请求，但是标记数据时发生错误：", err)
			}
		} else if queryUserInDatabaseStatus == "markRemove" && queryUserInDatabaseStatusCode == 200 {
			result.StatusCode = 418
			result.StatusString = "用户数据正在删除,请等待删除操作完成。删除操作需要等待四十八小时后才会执行，请耐心等待"
			context.JSON(http.StatusOK, result)
		} else if queryUserInDatabaseStatus == "markRemove" && queryUserInDatabaseStatusCode == 500 {
			result.StatusCode = 500
			result.StatusString = "用户数据删除请求已提交满四十八小时，但是API尝试删除用户失败，这是由于系统内尚未刷新该数据允许删除，API会自动重试直到系统允许删除，请等待"
			context.JSON(http.StatusOK, result)
		}
	}
}
