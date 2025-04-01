package types

// VersionRecord 表示一个键的历史版本记录
type VersionRecord struct {
	Version        int64  `json:"version"`
	Value          string `json:"value"`
	Revision       int64  `json:"revision"`
	CreateRevision int64  `json:"create_revision"`
	ModRevision    int64  `json:"mod_revision"`
}

// Options 定义查询历史版本的选项
type Options struct {
	// Endpoints 定义etcd服务器的地址列表
	Endpoints []string
	// Key 要查询历史版本的键
	Key string
	// Limit 限制返回的记录数量，0表示不限制
	Limit int
	// Order 排序方式，"asc"表示升序，"desc"表示降序
	Order string
	// Username etcd认证的用户名
	Username string
	// Password etcd认证的密码
	Password string
}