package types

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type User struct {
	ID             primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	FirstName      string             `json:"firstName"`
	LastName       string             `json:"lastName"`
	Email          string             `json:"email"`
	Password       string             `json:"password"`
	PhotoURL       string             `json:"photoURL"`
	BannerURL      string             `json:"bannerURL"`
	FollowersCount uint32             `json:"followersCount"`
	FollowingCount uint32             `json:"followingCount"`
	Bio            string             `json:"bio"`
	FollowedBy     []string           `json:"followed_by"`
	FollowedUsers  []string           `json:"followedUsers"`
	CreatedAt      time.Time          `json:"createdAt"`
}

type Post struct {
	ID            primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	UserID        primitive.ObjectID `json:"userID,omitempty" bson:"userID"`
	Caption       string             `json:"caption"`
	Images        []string           `json:"images"`
	Tags          []string           `json:"tags"`
	LikesCount    uint32             `json:"likesCount"`
	CommentsCount uint32             `json:"commentsCount"`
	LikedBy       []string           `json:"likedBy"`
	CommentedBy   []string           `json:"commentedBy"`
	CreatedAt     time.Time          `json:"createdAt"`
	UpdatedAt     time.Time          `json:"updatedAt"`
}
