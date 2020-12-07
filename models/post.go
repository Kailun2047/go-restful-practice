package models

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Post struct {
	ID        uint      `json:"id"`
	Title     string    `gorm:"size:255;not null;unique" json:"title"`
	Content   string    `gorm:"size:255;not null" json:"content"`
	UserID    uint      `json:"user_id"` // The other side of one-to-many association.
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (p *Post) Validate() error {
	if p.Title == "" {
		return fmt.Errorf("Title is a require field")
	}
	if p.Content == "" {
		return fmt.Errorf("Content is a require field")
	}
	if p.UserID < 1 {
		return fmt.Errorf("A valid user ID is required")
	}
	return nil
}

func (p *Post) SavePost(db *gorm.DB) (*Post, error) {
	// Make sure user ID is valid first.
	err := db.Debug().Model(&User{}).Where("id = ?", p.UserID).First(&User{}).Error
	if err != nil {
		return &Post{}, err
	}

	err = db.Debug().Model(&Post{}).Create(p).Error
	if err != nil {
		return &Post{}, err
	}
	return p, err
}

func (p *Post) FindAllPosts(db *gorm.DB) ([]Post, error) {
	posts := []Post{}
	err := db.Debug().Model(&Post{}).Find(posts).Error
	if err != nil {
		return []Post{}, err
	}
	return posts, err
}

func (p *Post) FindPostByID(db *gorm.DB, id uint) (*Post, error) {
	err := db.Debug().Model(&Post{}).Where("id = ?", id).First(p).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &Post{}, fmt.Errorf("Post with ID [%d] not found", id)
		}
	}
	return p, err
}

func (p *Post) UpdatePost(db *gorm.DB, id uint) (*Post, error) {
	db = db.Debug().Model(&Post{}).Where("id = ?", id).First(&Post{}).Updates(
		map[string]interface{}{
			"title":   p.Title,
			"content": p.Content,
		},
	)
	if err := db.Error; err != nil {
		return &Post{}, err
	}
	err := db.Debug().Model(&Post{}).Where("id = ", id).First(p).Error
	if err != nil {
		return &Post{}, err
	}
	return p, err
}

func (p *Post) DeletePost(db *gorm.DB, id uint) (int64, error) {
	db = db.Debug().Model(&Post{}).Where("id = ?", id).Delete(&Post{})
	if err := db.Error; err != nil {
		return 0, err
	}
	return db.RowsAffected, nil
}
