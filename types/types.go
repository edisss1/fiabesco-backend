package types

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type User struct {
	ID             primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	FirstName      string             `json:"firstName" bson:"firstName"`
	LastName       string             `json:"lastName" bson:"lastName"`
	Email          string             `json:"email" bson:"email"`
	Password       string             `json:"password" bson:"password"`
	Handle         string             `json:"handle"`
	Token          string             `json:"token" bson:"token"`
	PhotoURL       string             `json:"photoURL" bson:"photoURL"`
	BannerURL      string             `json:"bannerURL" bson:"bannerURL"`
	FollowersCount uint32             `json:"followersCount" bson:"followersCount"`
	FollowingCount uint32             `json:"followingCount" bson:"followingCount"`
	Bio            string             `json:"bio" bson:"bio"`
	FollowedBy     []string           `json:"followedBy" bson:"followedBy"`
	FollowedUsers  []string           `json:"followedUsers" bson:"followedUsers"`
	CreatedAt      time.Time          `json:"createdAt" bson:"createdAt"`
}

type Post struct {
	ID            primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	UserID        primitive.ObjectID `json:"userID,omitempty" bson:"userID"`
	UserFirstName string             `json:"userFirstName" bson:"userFirstName"`
	UserLastName  string             `json:"userLastName" bson:"userLastName"`
	UserPhotoURL  string             `json:"userPhotoURL" bson:"userPhotoURL"`
	UserHandle    string             `json:"userHandle" bson:"userHandle"`
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
