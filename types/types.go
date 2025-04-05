package types

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID             primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	FirstName      string             `json:"firstName"`
	LastName       string             `json:"lastName"`
	Email          string             `json:"email"`
	Password       string             `json:"password"`
	PhotoURL       string             `json:"photoURL"`
	BannerURL      string             `json:"bannerURL"`
	FollowersCount int16              `json:"followersCount"`
	FollowingCount int16              `json:"followingCount"`
	Bio            string             `json:"bio"`
	FollowedBy     []string           `json:"followed_by"`
	FollowedUsers  []string           `json:"followedUsers"`
}
