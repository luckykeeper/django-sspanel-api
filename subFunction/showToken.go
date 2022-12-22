// subFunction of vpnHelper
// Powered By Luckykeeper <luckykeeper@luckykeeper.site | https://luckykeeper.site>
package subFunction

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// 显示 Token
func ShowToken(VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName string) {
	db, _ := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName))
	defer db.Close()
	querySql := "SELECT * FROM token;"
	db.Exec(querySql)
	var token string
	queryResult := db.QueryRow(querySql).Scan(&token)
	if queryResult == sql.ErrNoRows {
		fmt.Println("当前系统内无 Token ，请先生成 Token !")
	} else {
		fmt.Println("当前系统内可用 Token 如下：")
		rows, _ := db.Query(querySql)
		defer rows.Close()
		for rows.Next() {
			rows.Scan(&token)
			fmt.Println(token)
		}
	}
}
