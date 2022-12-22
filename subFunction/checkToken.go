// subFunction of vpnHelper
// Powered By Luckykeeper <luckykeeper@luckykeeper.site | https://luckykeeper.site>
package subFunction

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// 鉴权，检查 Token
func CheckToken(tokenToCheck,
	VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName string) bool {
	db, _ := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName))
	defer db.Close()
	querySql := "SELECT * FROM token WHERE token='" + tokenToCheck + "';"
	db.Exec(querySql)
	var token string
	queryResult := db.QueryRow(querySql).Scan(&token)
	if queryResult == sql.ErrNoRows {
		// 错误的 Token
		return false
	} else {
		// Token 正确
		return true
	}
}
