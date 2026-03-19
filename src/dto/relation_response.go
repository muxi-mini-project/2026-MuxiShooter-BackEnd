package dto

import (
	models "MuXi/2026-MuxiShooter-Backend/models"
	"time"
)

type AchievementBrief struct {
	AchievementID   uint   `json:"achievement_id"`
	AchievementName string `json:"achievement_name"`
	Description     string `json:"description"`
}

type SkillBrief struct {
	SkillID     uint   `json:"skill_id"`
	SkillName   string `json:"skill_name"`
	Description string `json:"description"`
	SkillGroup  string `json:"skill_group"`
	PrqSkillID  uint   `json:"prq_skill_id"`
}

type ItemBrief struct {
	ItemID      uint   `json:"item_id"`
	ItemName    string `json:"item_name"`
	Description string `json:"description"`
}

type CardBrief struct {
	CardID      uint   `json:"card_id"`
	CardName    string `json:"card_name"`
	Description string `json:"description"`
}

type UserAchievementRelationData struct {
	IsComplete  bool             `json:"is_complete"`
	CompleteAt  *time.Time       `json:"complete_at,omitempty"`
	Claimed     bool             `json:"claimed"`
	ClaimedAt   *time.Time       `json:"claimed_at,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	Achievement AchievementBrief `json:"achievement"`
}

type UserSkillRelationData struct {
	IsComplete bool       `json:"is_complete"`
	CompleteAt *time.Time `json:"complete_at,omitempty"`
	SkillGrade uint       `json:"skill_grade"`
	Claimed    bool       `json:"claimed"`
	ClaimedAt  *time.Time `json:"claimed_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	Skill      SkillBrief `json:"skill"`
}

type UserItemRelationData struct {
	IsComplete bool       `json:"is_complete"`
	CompleteAt *time.Time `json:"complete_at,omitempty"`
	Claimed    bool       `json:"claimed"`
	ClaimedAt  *time.Time `json:"claimed_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	Item       ItemBrief  `json:"item"`
}

type UserCardRelationData struct {
	IsComplete bool       `json:"is_complete"`
	CompleteAt *time.Time `json:"complete_at,omitempty"`
	Claimed    bool       `json:"claimed"`
	ClaimedAt  *time.Time `json:"claimed_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	Card       CardBrief  `json:"card"`
}

type UserAchievementRelationPageData struct {
	List     []UserAchievementRelationData `json:"list"`
	Total    int64                         `json:"total"`
	Page     int                           `json:"page"`
	PageSize int                           `json:"page_size"`
}

type UserSkillRelationPageData struct {
	List     []UserSkillRelationData `json:"list"`
	Total    int64                   `json:"total"`
	Page     int                     `json:"page"`
	PageSize int                     `json:"page_size"`
}

type UserItemRelationPageData struct {
	List     []UserItemRelationData `json:"list"`
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
}

type UserCardRelationPageData struct {
	List     []UserCardRelationData `json:"list"`
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
}

type CommonRelationResourceData struct {
	ResourceID   uint   `json:"resource_id"`
	ResourceName string `json:"resource_name"`
	Description  string `json:"description"`
	SkillGroup   string `json:"skill_group,omitempty"`
	PrqSkillID   uint   `json:"prq_skill_id,omitempty"`
}

type CommonUserRelationData struct {
	IsComplete bool                       `json:"is_complete"`
	CompleteAt *time.Time                 `json:"complete_at,omitempty"`
	SkillGrade uint                       `json:"skill_grade,omitempty"`
	Claimed    bool                       `json:"claimed"`
	ClaimedAt  *time.Time                 `json:"claimed_at,omitempty"`
	CreatedAt  time.Time                  `json:"created_at"`
	UpdatedAt  time.Time                  `json:"updated_at"`
	Resource   CommonRelationResourceData `json:"resource"`
}

type CommonUserRelationPageData struct {
	List     []CommonUserRelationData `json:"list"`
	Total    int64                    `json:"total"`
	Page     int                      `json:"page"`
	PageSize int                      `json:"page_size"`
}

func BuildUserAchievementRelationList(records []models.UserAchievement) []UserAchievementRelationData {
	result := make([]UserAchievementRelationData, 0, len(records))
	for _, record := range records {
		result = append(result, UserAchievementRelationData{
			IsComplete: record.IsComplete,
			CompleteAt: record.CompleteAt,
			Claimed:    record.Claimed,
			ClaimedAt:  record.ClaimedAt,
			CreatedAt:  record.CreatedAt,
			UpdatedAt:  record.UpdatedAt,
			Achievement: AchievementBrief{
				AchievementID:   record.Achievement.ID,
				AchievementName: record.Achievement.Name,
				Description:     record.Achievement.Description,
			},
		})
	}
	return result
}

func BuildUserSkillRelationList(records []models.UserSkill) []UserSkillRelationData {
	result := make([]UserSkillRelationData, 0, len(records))
	for _, record := range records {
		result = append(result, UserSkillRelationData{
			IsComplete: record.IsComplete,
			CompleteAt: record.CompleteAt,
			SkillGrade: record.SkillGrade,
			Claimed:    record.Claimed,
			ClaimedAt:  record.ClaimedAt,
			CreatedAt:  record.CreatedAt,
			UpdatedAt:  record.UpdatedAt,
			Skill: SkillBrief{
				SkillID:     record.Skill.ID,
				SkillName:   record.Skill.Name,
				Description: record.Skill.Description,
				SkillGroup:  record.Skill.SkillGroup,
				PrqSkillID:  record.Skill.PrqSkillId,
			},
		})
	}
	return result
}

func BuildUserItemRelationList(records []models.UserItem) []UserItemRelationData {
	result := make([]UserItemRelationData, 0, len(records))
	for _, record := range records {
		result = append(result, UserItemRelationData{
			IsComplete: record.IsComplete,
			CompleteAt: record.CompleteAt,
			Claimed:    record.Claimed,
			ClaimedAt:  record.ClaimedAt,
			CreatedAt:  record.CreatedAt,
			UpdatedAt:  record.UpdatedAt,
			Item: ItemBrief{
				ItemID:      record.Item.ID,
				ItemName:    record.Item.Name,
				Description: record.Item.Description,
			},
		})
	}
	return result
}

func BuildUserCardRelationList(records []models.UserCard) []UserCardRelationData {
	result := make([]UserCardRelationData, 0, len(records))
	for _, record := range records {
		result = append(result, UserCardRelationData{
			IsComplete: record.IsComplete,
			CompleteAt: record.CompleteAt,
			Claimed:    record.Claimed,
			ClaimedAt:  record.ClaimedAt,
			CreatedAt:  record.CreatedAt,
			UpdatedAt:  record.UpdatedAt,
			Card: CardBrief{
				CardID:      record.Card.ID,
				CardName:    record.Card.Name,
				Description: record.Card.Description,
			},
		})
	}
	return result
}
