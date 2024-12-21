package models

import (
	"time"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	db.Set("gorm:table_options", "ENGINE=InnoDB")
	if err := db.AutoMigrate(
		&User{},
		&Template{},
		&Channel{},
		&Router{},
		&Audit{}); err != nil {
		return err
	}
	return nil
}

type Model struct {
	ID        uint           `gorm:"primary_key" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
