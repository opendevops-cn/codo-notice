package models

type Audit struct {
	Model
	CreatedBy      string `gorm:"type:varchar(128)" json:"created_by"`
	UpdatedBy      string `gorm:"type:varchar(128)" json:"updated_by"`
	ResourceType   string `gorm:"type:varchar(32);index:idx_type;not null;comment:资源类型" json:"resource_type"`
	ResourceName   string `gorm:"type:varchar(128);comment:资源目标/名称" json:"resource_name"`
	ResourceTarget string `gorm:"type:varchar(2048);comment:资源目标/实例" json:"resource_target"`
	Action         string `gorm:"type:varchar(32);index:idx_action;not null;comment:操作类型" json:"action"`
	Status         string `gorm:"type:varchar(16);default:success;comment:状态" json:"status"` // success/failed
	Detail         []byte `gorm:"type:mediumtext;comment:详情" json:"detail"`
}
