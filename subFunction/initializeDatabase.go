// subFunction of vpnHelper
// Powered By Luckykeeper <luckykeeper@luckykeeper.site | https://luckykeeper.site>
package subFunction

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

const (
	sql_initialize_table_token      = `CREATE TABLE token (token TEXT NOT NULL)WITH (OIDS=FALSE);`
	sql_initialize_table_vpnhelper  = `CREATE TABLE vpnHelper (username TEXT NOT NULL UNIQUE,status TEXT NOT NULL,statusCode int,loginpasswd TEXT,connpasswd TEXT,requesttime TEXT)WITH (OIDS=FALSE);`
	sql_initialize_table_inviteCode = `CREATE TABLE inviteCode (inviteCode TEXT NOT NULL UNIQUE)WITH (OIDS=FALSE);`
	sql_initialize_table_nodeinfo   = `CREATE TABLE nodeinfo (sequence TEXT NOT NULL UNIQUE,name TEXT NOT NULL UNIQUE,nodetype TEXT NOT NULL,tips TEXT NOT NULL,ip TEXT NOT NULL,method TEXT NOT NULL,port TEXT NOT NULL)WITH (OIDS=FALSE);`
)

// PGSQL 初始化
func InitializeDatabase(VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName string) {
	db, _ := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName))
	defer db.Close()
	var checkInitializeStatus int
	db.QueryRow(`select count(*) from pg_statio_user_tables where relname='token';`).Scan(&checkInitializeStatus)
	// 不存在就创建并初始化，存在则跳过
	if checkInitializeStatus != 1 {
		//初始化
		log.Println("初始化数据库中……")
		db, _ := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s sslmode=disable",
			VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword))
		defer db.Close()
		db.Exec(fmt.Sprintf((`CREATE DATABASE %s;`), VpnHelperDBName))
		log.Println("初始化数据库完成")
		log.Println("初始化数据表中……")
		db1, _ := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName))
		defer db1.Close()
		db1.Exec(sql_initialize_table_token)
		db1.Exec(sql_initialize_table_vpnhelper)
		db1.Exec(sql_initialize_table_inviteCode)
		db1.Exec(sql_initialize_table_nodeinfo)
		log.Println("初始化数据表完成")
	} else {
		log.Println("数据库连接成功")
	}
}

// 初始化表 inviteCode
func initializeTableInviteCode(VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName string) {
	db, _ := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		VpnHelperDBAddress, VpnHelperDBPort, VpnHelperDBUsername, VpnHelperDBPassword, VpnHelperDBName))
	defer db.Close()
	db.Exec("DROP TABLE inviteCode;")
	db.Exec(sql_initialize_table_inviteCode)
}
