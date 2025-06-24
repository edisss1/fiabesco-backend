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
	PhotoURL       string             `json:"photoURL" bson:"photoURL"`
	BannerURL      string             `json:"bannerURL" bson:"bannerURL"`
	FollowersCount uint32             `json:"followersCount" bson:"followersCount"`
	FollowingCount uint32             `json:"followingCount" bson:"followingCount"`
	Bio            string             `json:"bio" bson:"bio"`
	FollowedBy     []string           `json:"followedBy" bson:"followedBy"`
	FollowedUsers  []string           `json:"followedUsers" bson:"followedUsers"`
	CreatedAt      time.Time          `json:"createdAt" bson:"createdAt"`
	Settings       *Settings          `json:"settings" bson:"settings"`
}

type Post struct {
	ID            primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	UserID        primitive.ObjectID `json:"userID,omitempty" bson:"userID"`
	Caption       string             `json:"caption"`
	Files         []string           `json:"images"`
	Tags          []string           `json:"tags"`
	LikesCount    uint32             `json:"likesCount" bson:"likesCount"`
	CommentsCount uint32             `json:"commentsCount" bson:"commentsCount"`
	LikedBy       []string           `json:"likedBy" bson:"likedBy"`
	CommentedBy   []string           `json:"commentedBy" bson:"commentedBy"`
	CreatedAt     time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt     time.Time          `json:"updatedAt" bson:"updatedAt"`
}

type Message struct {
	ID             primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	ConversationID primitive.ObjectID `json:"conversationID,omitempty" bson:"conversationID"`
	SenderID       primitive.ObjectID `json:"senderID" bson:"senderID"`
	Content        string             `json:"content"`
	Files          []string           `json:"uploads"`
	Read           bool               `json:"read"`
	CreatedAt      time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt      time.Time          `json:"updatedAt" bson:"updatedAt"`
	IsEdited       bool               `json:"isEdited" bson:"isEdited"`
}

type Conversation struct {
	ID           primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty"`
	Participants []primitive.ObjectID `json:"participants" bson:"participants"`
	Names        []string             `json:"names"`
	IsGroup      bool                 `json:"isGroup" bson:"isGroup"`
	Name         string               `json:"name"`
	LastMessage  Message              `json:"lastMessage" bson:"lastMessage"`
	CreatedAt    time.Time            `json:"createdAt" bson:"createdAt"`
	UpdatedAt    time.Time            `json:"updatedAt" bson:"updatedAt"`
}

type Repost struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"id"`
	PostID    primitive.ObjectID `json:"postID" bson:"postID"`
	UserID    primitive.ObjectID `json:"userID" bson:"userID"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
}

type Like struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	PostID    primitive.ObjectID `json:"postID" bson:"postID"`
	UserID    primitive.ObjectID `json:"userID" bson:"userID"`
	UserName  string             `json:"userName" bson:"userName"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
}

type Comment struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	PostID    primitive.ObjectID `json:"postID" bson:"postID"`
	UserID    primitive.ObjectID `json:"userID" bson:"userID"`
	Content   string             `json:"content"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
}

type Settings struct {
	Theme    string `json:"theme" bson:"theme"`
	Language string `json:"language"`
}

type Follow struct {
	FollowerID  primitive.ObjectID `json:"followerID" bson:"followerID"` // user that follows
	FollowingID primitive.ObjectID `json:"followedID" bson:"followedID"` // user that is being followed
}

type Block struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	UserID    primitive.ObjectID `json:"userID" bson:"userID"`       // who is blocking
	BlockedID primitive.ObjectID `json:"blockedID" bson:"blockedID"` // who is being blocked
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
}

type Portfolio struct {
	ID          string               `json:"_id,omitempty" bson:"_id,omitempty"`
	UserID      string               `json:"userID" bson:"userID"`
	AllowEmails bool                 `json:"allowEmails" bson:"allowEmails"`
	About       string               `json:"about"`
	Projects    []PortfolioProject   `json:"projects"`
	Appearance  PortfolioAppearance  `json:"appearance"`
	ContactInfo PortfolioContactInfo `json:"contactInfo" bson:"contactInfo"`
}

type PortfolioProject struct {
	Img   string `json:"img"`
	Title string `json:"title"`
	Link  string `json:"link"`
}

type PortfolioAppearance struct {
	TextColor    string `json:"textColor" bson:"textColor"`
	BgColor      string `json:"bgColor" bson:"bgColor"`
	PrimaryColor string `json:"primaryColor" bson:"primaryColor"`
}

type PortfolioContactInfo struct {
	Email                 string `json:"email"`
	BehanceProfileLink    string `json:"behanceProfileLink" bson:"behanceProfileLink"`
	DribbbleProfileLink   string `json:"dribbbleProfileLink" bson:"dribbbleProfileLink"`
	PinterestProfileLink  string `json:"pinterestProfileLink" bson:"pinterestProfileLink"`
	ArtStationProfileLink string `json:"artStationProfileLink" bson:"artStationProfileLink"`
}
