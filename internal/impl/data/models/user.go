package models

type User struct {
	Model
	Username   string `gorm:"type:varchar(128);comment:名称" json:"username" mapstructure:"username"`
	Nickname   string `gorm:"type:varchar(128);comment:昵称" json:"nickname" mapstructure:"nickname"`
	UserId     string `gorm:"type:varchar(128);unique;not null;comment:用户标识" json:"user_id" mapstructure:"user_id"`
	DepId      string `gorm:"type:varchar(2048);comment:部门标识" json:"dep_id" mapstructure:"dep_id"`
	Dep        string `gorm:"type:varchar(2048);comment:部门" json:"dep" mapstructure:"dep"`
	Manager    string `gorm:"type:varchar(128);comment:管理者" json:"manager" mapstructure:"manager"`
	Avatar     string `gorm:"type:varchar(1024);comment:头像" json:"avatar" mapstructure:"avatar"`
	Active     bool   `gorm:"type:tinyint(1);default:true;comment:是否激活" json:"active" mapstructure:"active"`
	Tel        string `gorm:"type:varchar(32);comment:手机" json:"tel" mapstructure:"tel"`
	Email      string `gorm:"type:varchar(128);comment:邮件" json:"email" mapstructure:"email"`
	DataSource []byte `gorm:"type:mediumtext;comment:源数据" json:"data_source" mapstructure:"data_source"`
	Disable    bool   `gorm:"type:tinyint(1);default:false;comment:是否禁用" json:"disable"`
	DdID       string `gorm:"type:varchar(128);comment:钉钉ID" json:"dd_id" mapstructure:"dd_id"`
	FsID       string `gorm:"type:varchar(128);comment:飞书ID" json:"fs_id" mapstructure:"fs_id"`
	WxID       string `gorm:"type:varchar(128);comment:企业微信ID" json:"wx_id" mapstructure:"wx_id"`
}
