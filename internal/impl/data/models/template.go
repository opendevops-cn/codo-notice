package models

type Template struct {
	Model
	CreatedBy string `gorm:"type:varchar(128)" json:"created_by"`
	UpdatedBy string `gorm:"type:varchar(128)" json:"updated_by"`
	Name      string `gorm:"size:256;index:idx_name;unique;not null;comment:模版名称" json:"name"`
	Content   string `gorm:"type:text;comment:模版信息" json:"content"`
	Type      string `gorm:"type:varchar(16);index:idx_type;not null;default:default;comment:类型" json:"type"`
	Use       string `gorm:"type:varchar(45);not null;default:default;comment:用途" json:"use"`
	// Default     uint   		`gorm:"type:int(11);default:0" json:"default"`	// 是否默认模版
	Default string `gorm:"type:varchar(16);default:no" json:"default"` // yes/no
}
