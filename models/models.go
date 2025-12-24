package models

// @externalDocs description="GORM Documentation" url="https://gorm.io/docs/"
import (
	"time"
)

// @summary		通用响应结构体
// @description	通用响应结构体
// @param			code	int		"状态码"
// @param			message	string	"消息"
// @param			data	object	"数据，类型根据接口变化"
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// @summary		用户信息
// @description	用户结构体
type User struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"user_id"`
	Username  string    `gorm:"unique;not null" json:"username"`
	Password  string    `gorm:"not null" json:"-"`
	Group     string    `gorm:"default:'user'" json:"group"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// @summary		书籍信息
// @description	书籍信息
type Book struct {
	ID      uint   `gorm:"primaryKey;autoIncrement" json:"book_id"`
	Title   string `json:"title"`
	Author  string `json:"author"`
	Summary string `json:"summary"`
	//本来不打算弄封面啥的，想想还是弄一下吧
	//当练手了qwq
	CoverPath    string    `json:"cover_path"`
	InitialStock int       `json:"initial_stock" gorm:"default:0" binding:"gte=0"`
	Stock        int       `json:"stock" gorm:"default:0" binding:"gte=0"`
	TotalStock   int       `json:"total_stock" gorm:"default:0" binding:"gte=0"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// @summary 借阅记录
// @description 借阅记录
// @property id uint "记录ID"
// @property created_at string "创建时间 (RFC3339)"
// @property updated_at string "更新时间 (RFC3339)"
// @property deleted_at string "删除时间 (RFC3339) 可为空 不为空时被删除"
// @property user_id uint "用户ID"
// @property book_id uint "图书ID"
// @property borrow_date string "借出时间 (RFC3339)"
// @property return_date string "归还时间 (RFC3339) 未归还时为空"
// @property status string "状态: borrowed or returned"
type BorrowRecord struct {
	ID        uint       `gorm:"primaryKey;autoIncrement" json:"record_id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"` //注意要用指针
	//这个删除好像只是查找时候不出现，也就是不是真删
	//用指针好像也是这个原因，nil就是没删
	UserID   uint       `json:"user_id"`
	BookID   uint       `json:"book_id"`
	BorrowAt time.Time  `json:"borrow_at"`
	ReturnAt *time.Time `json:"return_at"`
	//同理
	Status string `json:"status"`
	//borrowed or returned
	User User `json:"user,omitempty" swaggerignore:"true"`
	Book Book `json:"book,omitempty" swaggerignore:"true"`
	//方便调用
}
