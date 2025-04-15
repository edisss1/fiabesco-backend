package types

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type User struct {
	ID             primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty"`
	FirstName      string               `json:"firstName" bson:"firstName"`
	LastName       string               `json:"lastName" bson:"lastName"`
	Email          string               `json:"email" bson:"email"`
	Password       string               `json:"password" bson:"password"`
	Handle         string               `json:"handle"`
	Token          string               `json:"token" bson:"token"`
	PhotoURL       string               `json:"photoURL" bson:"photoURL"`
	BannerURL      string               `json:"bannerURL" bson:"bannerURL"`
	FollowersCount uint32               `json:"followersCount" bson:"followersCount"`
	FollowingCount uint32               `json:"followingCount" bson:"followingCount"`
	Bio            string               `json:"bio" bson:"bio"`
	FollowedBy     []string             `json:"followedBy" bson:"followedBy"`
	FollowedUsers  []string             `json:"followedUsers" bson:"followedUsers"`
	CreatedAt      time.Time            `json:"createdAt" bson:"createdAt"`
	BlockedUsers   []primitive.ObjectID `json:"blockedUsers,omitempty" bson:"blockedUsers,omitempty"`
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

type Message struct {
	ID             primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	ConversationID primitive.ObjectID `json:"conversationID,omitempty" bson:"conversationID"`
	SenderID       primitive.ObjectID `json:"senderID" bson:"senderID"`
	Content        string             `json:"content"`
	Files          []string           `json:"files"`
	Read           bool               `json:"read"`
	CreatedAt      time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt      time.Time          `json:"updatedAt" bson:"updatedAt"`
}

type Conversation struct {
	ID           primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty"`
	Participants []primitive.ObjectID `json:"participants" bson:"participants"`
	Names        []string             `json:"names"`
	IsGroup      bool                 `json:"isGroup" bson:"isGroup"`
	Name         string               `json:"name"`
	LastMessage  string               `json:"lastMessage" bson:"lastMessage"`
	CreatedAt    time.Time            `json:"createdAt" bson:"createdAt"`
	UpdatedAt    time.Time            `json:"updatedAt" bson:"updatedAt"`
}
