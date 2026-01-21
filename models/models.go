package models

//	@externalDocs	description="GORM Documentation" url="https://gorm.io/docs/"
import (
	"time"
)

type User struct {
	ID                 uint              `gorm:"primaryKey;autoIncrement" json:"user_id"`
	TokenVersion       uint64            `gorm:"default:0" json:"token_version"`
	Username           string            `gorm:"unique;not null" json:"username"`
	UsernameUpdatedAt  *time.Time        `json:"username_updated_at"`
	Password           string            `gorm:"not null" json:"-"`
	PasswordUpdatedAt  *time.Time        `json:"password_updated_at"`
	Group              string            `gorm:"default:'user'" json:"group"`
	HeadImagePath      string            `json:"head_image_path"`
	HeadImageUpdatedAt *time.Time        `json:"head_image_updated_at"`
	StrengthCoin       uint              `gorm:"default:0" json:"strength_coin"`
	SelectCoin         uint              `gorm:"default:0" json:"select_coin"`
	UserAchievements   []UserAchievement `gorm:"foreignKey:UserID" json:"-"`
	UserSkills         []UserSkill       `gorm:"foreignKey:UserID" json:"-"`
	UserCards          []UserCard        `gorm:"foreignKey:UserID" json:"-"`
	UserItems          []UserItem        `gorm:"foreignKey:UserID" json:"-"`
	CreatedAt          time.Time         `json:"created_at"`
	UpdatedAt          time.Time         `json:"updated_at"`
}

type Achievement struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"achievement_id"`
	Name        string    `gorm:"unique;not null" json:"achievement_name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UserAchievement struct {
	UserID        uint       `gorm:"primaryKey" json:"-"`
	AchievementID uint       `gorm:"primaryKey" json:"-"`
	IsComplete    bool       `gorm:"default:false" json:"is_complete"`
	CompleteAt    *time.Time `json:"complete_at,omitempty"`
	Claimed       bool       `gorm:"default:false" json:"claimed"`
	ClaimedAt     *time.Time `json:"claimed_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`

	//关联关系
	User        User        `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Achievement Achievement `gorm:"foreignKey:AchievementID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type Skill struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"skill_id"`
	Name        string    `gorm:"unique;not null" json:"skill_name"`
	Description string    `json:"description"`
	SkillGroup  string    `json:"skill_group"` //Front End Products Design Operations Apple Android
	PrqSkillId  uint      `json:"prq_skill_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UserSkill struct {
	UserID     uint       `gorm:"primaryKey" json:"-"`
	SkillID    uint       `gorm:"primaryKey" json:"-"`
	IsComplete bool       `gorm:"default:false" json:"is_complete"`
	CompleteAt *time.Time `json:"complete_at,omitempty"`
	SkillGrade uint       `gorm:"default:0" json:"skill_grade"`
	Claimed    bool       `gorm:"default:false" json:"claimed"`
	ClaimedAt  *time.Time `json:"claimed_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`

	//关联关系
	User  User  `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Skill Skill `gorm:"foreignKey:SkillID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type Card struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"skill_id"`
	Name        string    `gorm:"unique;not null" json:"skill_name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UserCard struct {
	UserID     uint       `gorm:"primaryKey" json:"-"`
	CardID     uint       `gorm:"primaryKey" json:"-"`
	IsComplete bool       `gorm:"default:false" json:"is_complete"`
	CompleteAt *time.Time `json:"complete_at,omitempty"`
	Claimed    bool       `gorm:"default:false" json:"claimed"`
	ClaimedAt  *time.Time `json:"claimed_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`

	//关联关系
	User User `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Card Card `gorm:"foreignKey:CardID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type Item struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"skill_id"`
	Name        string    `gorm:"unique;not null" json:"skill_name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UserItem struct {
	UserID     uint       `gorm:"primaryKey" json:"-"`
	ItemID     uint       `gorm:"primaryKey" json:"-"`
	IsComplete bool       `gorm:"default:false" json:"is_complete"`
	CompleteAt *time.Time `json:"complete_at,omitempty"`
	Claimed    bool       `gorm:"default:false" json:"claimed"`
	ClaimedAt  *time.Time `json:"claimed_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`

	User User `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Item Item `gorm:"foreignKey:ItemID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
