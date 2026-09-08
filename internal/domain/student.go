package domain

import "time"

type Student struct {
	ID        string    `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	ParentID  string    `json:"parent_id" gorm:"type:uuid;not null"`
	Name      string    `json:"name" gorm:"type:varchar(255);not null"`
	Age       int       `json:"age" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	Parent Parent `json:"parent" gorm:"foreignKey:ParentID"`
}
