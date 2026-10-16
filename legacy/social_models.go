package models

import (
	"time"

	"gorm.io/gorm"
)

// User model
type User struct {
	gorm.Model
	Username     string `gorm:"uniqueIndex;not null"`
	Email        string `gorm:"uniqueIndex;not null"`
	Password     string `gorm:"not null"`
	Age          int
	Bio          string
	AvatarURL    string
	ProfileType  string // individual/couple
	Gender       string
	Interests    string
	LocationLat  float64
	LocationLong float64
	IsPremium    bool `gorm:"default:false"`
	IsPrivate    bool `gorm:"default:false"`

	Photos   []Photo   `gorm:"foreignKey:UserID"`
	Posts    []Post    `gorm:"foreignKey:UserID"`
	Comments []Comment `gorm:"foreignKey:UserID"`

	// Interações
	Likes          []Like  `gorm:"foreignKey:UserID"`
	SentBlocks     []Block `gorm:"foreignKey:BlockerID"`
	ReceivedBlocks []Block `gorm:"foreignKey:BlockedID"`

	// Comunicação
	Chats []Chat `gorm:"many2many:user_chats;"`
}

// Photo model
type Photo struct {
	gorm.Model
	UserID    uint
	URL       string `gorm:"not null"`
	IsPrivate bool   `gorm:"default:false"`

	Likes    []PhotoLike `gorm:"foreignKey:PhotoID"`
	Comments []Comment   `gorm:"foreignKey:PhotoID"`
}

// Post model
type Post struct {
	gorm.Model
	UserID   uint
	Text     string
	ImageURL string

	Likes    []PostLike `gorm:"foreignKey:PostID"`
	Comments []Comment  `gorm:"foreignKey:PostID"`
}

// Comment model
type Comment struct {
	gorm.Model
	UserID uint
	Text   string `gorm:"not null"`

	// Pode pertencer a uma Photo ou a um Post
	PhotoID *uint
	PostID  *uint

	Likes []CommentLike `gorm:"foreignKey:CommentID"`
}

// Like (User to User)
type Like struct {
	gorm.Model
	UserID  uint `gorm:"uniqueIndex:idx_user_liked"`
	LikedID uint `gorm:"uniqueIndex:idx_user_liked"`
}

// PostLike
type PostLike struct {
	gorm.Model
	UserID uint `gorm:"uniqueIndex:idx_post_like"`
	PostID uint `gorm:"uniqueIndex:idx_post_like"`
}

// PhotoLike
type PhotoLike struct {
	gorm.Model
	UserID  uint `gorm:"uniqueIndex:idx_photo_like"`
	PhotoID uint `gorm:"uniqueIndex:idx_photo_like"`
}

// CommentLike
type CommentLike struct {
	gorm.Model
	UserID    uint `gorm:"uniqueIndex:idx_comment_like"`
	CommentID uint `gorm:"uniqueIndex:idx_comment_like"`
}

// Block
type Block struct {
	gorm.Model
	BlockerID uint `gorm:"uniqueIndex:idx_block"`
	BlockedID uint `gorm:"uniqueIndex:idx_block"`
}

// Chat
type Chat struct {
	gorm.Model
	Users    []User    `gorm:"many2many:user_chats;"`
	Messages []Message `gorm:"foreignKey:ChatID"`
}

// Message
type Message struct {
	gorm.Model
	ChatID   uint
	SenderID uint
	Content  string
	ImageURL string
	SentAt   time.Time `gorm:"autoCreateTime"`
}
