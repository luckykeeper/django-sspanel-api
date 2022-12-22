// subFunction of vpnHelper
// Powered By Luckykeeper <luckykeeper@luckykeeper.site | https://luckykeeper.site>
package subFunction

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"

	_ "github.com/lib/pq"
)

// 新增 Token
func GenerateToken(VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName string) {
	rand.Seed(time.Now().UnixNano())
	uLen := 32
	b := make([]byte, uLen)
	rand.Read(b)
	rand_str := hex.EncodeToString(b)[0:uLen]
	fmt.Println("新 Token :", rand_str)
	db, _ := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName))
	defer db.Close()
	insertSql := "INSERT INTO token (token) values ('" + rand_str + "');"
	db.Exec(insertSql)
	fmt.Println("新增 Token 完成!")
}
