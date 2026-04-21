// package base 定义了基础的数据结构和相关操作
package base

// User 结构体定义了用户在数据库中的表结构
// 该结构体使用了 gorm 标签来指定数据库表的列名、约束条件等信息，同时使用 json 标签来指定 JSON 序列化时的字段名
type User struct {
	// Id 是用户表的主键，使用自增方式生成
	Id int64 `gorm:"column:id;AUTO_INCREMENT;primary_key" json:"id"`
	// Address 是用户的地址，该字段不能为空
	Address string `gorm:"column:address;NOT NULL" json:"address"`
	// IsAllowed 表示用户是否被允许访问，默认值为 0（即 false），该字段不能为空
	IsAllowed bool `gorm:"column:is_allowed;default:0;NOT NULL" json:"is_allowed"`
	// IsSigned 表示用户是否已签名，默认值为 0（即 false）
	IsSigned bool `gorm:"column:is_signed;default:0" json:"is_signed"`
	// CreateTime 是用户记录的创建时间，使用大整数类型，精确到毫秒，会自动记录创建时间
	CreateTime int64 `json:"create_time" gorm:"column:create_time;type:bigint(20);autoCreateTime:milli;comment:创建时间"`
	// UpdateTime 是用户记录的更新时间，使用大整数类型，精确到毫秒，会自动记录更新时间
	UpdateTime int64 `json:"update_time" gorm:"column:update_time;type:bigint(20);autoUpdateTime:milli;comment:更新时间"`
}

// UserTableName 返回用户表在数据库中的名称
// 返回值为字符串 "ob_user"
func UserTableName() string {
	return "ob_user"
}
