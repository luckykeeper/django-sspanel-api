// model of vpnHelper
// Powered By Luckykeeper <luckykeeper@luckykeeper.site | https://luckykeeper.site>
package model

// 从 OA 侧传递的用户模型
type UserModel struct {
	UserName string `form:"username"` // 用户名（学工号）
	Token    string `form:"token"`    // 鉴权 Token
}

// API 向 OA 返回的操作结果信息
type ResultMsg struct {
	StatusCode   int    `form:"statusCode"` // 结果码 （初次提交成功200，已提交正在处理102，用户存在不能重复操作201，Token错误401）
	StatusString string `form:"StatusString"`
}

// 管理面板 - 获取邀请码
type InviteCode struct {
	InviteCode string `form:"InviteCode"` // 邀请码
	CreateTime string `form:"CreateTime"` // 创建时间
	CodeUsed   string `form:"CodeUsed"`   // 是否使用
	CodeType   string `form:"CodeType"`   // 类型（是否公开）
}

// 管理面板 - 获取节点信息（接口返回共用）
type NodeInfo struct {
	Sequence string `form:"Sequence"` // 顺序
	Name     string `form:"Name"`     // 服务器名称
	NodeType string `form:"NodeType"` // 节点类型
	Tips     string `form:"Tips"`     // 节点说明
	Ip       string `form:"Ip"`       // 服务器 IP
	Method   string `form:"Method"`   // 加密类型（SS）
	Port     string `form:"Port"`     // 端口

	// 接口返回用
	ConnInfo string `form:"ConnInfo"` // 接入信息
}

// 查询用户连接信息接口
type ConnInfoByUserName struct {
	ResultMsg []ResultMsg
	NodeInfos []NodeInfo
}
