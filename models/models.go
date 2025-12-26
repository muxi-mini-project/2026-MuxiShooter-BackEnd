package models

//	@externalDocs	description="GORM Documentation" url="https://gorm.io/docs/"
import (
	"time"
)

// @description	用户信息
type User struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"user_id"`
	Username  string    `gorm:"unique;not null" json:"username"`
	Password  string    `gorm:"not null" json:"-"`
	Group     string    `gorm:"default:'user'" json:"group"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// @description	书籍信息
type Book struct {
	//图书ID minimum(1)
	ID uint `gorm:"primaryKey;autoIncrement" json:"book_id"`
	//书名
	Title string `json:"title"`
	//作者
	Author string `json:"author"`
	//简介
	Summary string `json:"summary"`
	//封面路径
	CoverPath string `json:"cover_path"`
	//初始库存
	InitialStock int `json:"initial_stock" gorm:"default:0" binding:"gte=0"`
	//现有库存
	Stock int `json:"stock" gorm:"default:0" binding:"gte=0"`
	//总库存
	TotalStock int `json:"total_stock" gorm:"default:0" binding:"gte=0"`
	//创建时间 (RFC3339)
	CreatedAt time.Time `json:"created_at"`
	//更新时间 (RFC3339)
	UpdatedAt time.Time `json:"updated_at"`
}

// @description	借阅记录
type BorrowRecord struct {
	//记录ID minimum(1)
	ID uint `gorm:"primaryKey;autoIncrement" json:"record_id"`
	//创建时间 (RFC3339)
	CreatedAt time.Time `json:"created_at"`
	//更新时间 (RFC3339)
	UpdatedAt time.Time `json:"updated_at"`
	//删除时间 (RFC3339)
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`
	//用户ID minimum(1)
	UserID uint `json:"user_id"`
	//图书ID minimum(1)
	BookID uint `json:"book_id"`
	//借书时间 (RFC3339)
	BorrowAt time.Time `json:"borrow_at"`
	//归还时间 (RFC3339)
	ReturnAt *time.Time `json:"return_at"`
	//borrowed or returned
	Status string `json:"status"`
	//用户信息
	User User `json:"-" swaggerignore:"true"`
	//书籍信息
	Book Book `json:"-" swaggerignore:"true"`
}
