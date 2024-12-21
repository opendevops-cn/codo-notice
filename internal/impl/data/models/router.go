package models

type Router struct {
	Model
	CreatedBy   string `gorm:"type:varchar(128)" json:"created_by"`
	UpdatedBy   string `gorm:"type:varchar(128)" json:"updated_by"`
	Name        string `gorm:"type:varchar(256);unique;comment:名称" json:"name"`
	Description string `gorm:"type:varchar(1024);comment:描述" json:"description"`
	// Active 		bool   		`gorm:"type:tinyint(1);comment:是否激活" json:"active"`
	Status    string `gorm:"type:varchar(16);default:yes;comment:是否激活" json:"status"` // yes/no
	ChannelId uint   `gorm:"type:int(11);default:0" json:"channel_id"`
	Condition []byte `gorm:"type:mediumtext;comment:触发条件" json:"condition_list"`
}
