package models

type Channel struct {
	Model
	CreatedBy string   `gorm:"type:varchar(128)" json:"created_by"`
	UpdatedBy string   `gorm:"type:varchar(128)" json:"updated_by"`
	Name      string   `gorm:"type:varchar(256);unique;comment:名称" json:"name"`
	Use       string   `gorm:"type:varchar(45);not null;default:default;comment:用途" json:"use"`
	User      []string `gorm:"serializer:json;comment:用户" json:"user"`
	Group     []uint   `gorm:"serializer:json;comment:用户组" json:"-"`
	/// --
	ContactPoints []byte `gorm:"type:mediumtext" json:"contact_points"`
	CustomItems   []byte `gorm:"type:mediumtext" json:"custom_items"`
	// 指定字段重写规则 / 一般用于短信电话 message
	DefaultRule map[string]string `gorm:"serializer:json" json:"-"`
}
