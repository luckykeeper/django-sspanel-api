// subFunction of vpnHelper
// Powered By Luckykeeper <luckykeeper@luckykeeper.site | https://luckykeeper.site>
package subFunction

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// 删除 Token
func DeleteToken(tokenToDelete,
	VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName string) {
	db, _ := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName))
	defer db.Close()
	querySql := "SELECT * FROM token WHERE token='" + tokenToDelete + "';"
	db.Exec(querySql)
	var token string
	queryResult := db.QueryRow(querySql).Scan(&token)
	if queryResult == sql.ErrNoRows {
		fmt.Println("系统内无此 Token:" + tokenToDelete + " ，请输入正确的 Token !")
	} else {
		deleteSql := "DELETE FROM token WHERE token='" + tokenToDelete + "';"
		db.Exec(deleteSql)
		fmt.Println("已删除 Token:", tokenToDelete)
	}
}
