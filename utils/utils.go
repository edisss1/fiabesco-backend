package utils

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"math/rand"
)

var baseImgURL = "http://localhost:3000/images"

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

type PipelineBuilder struct {
	stages mongo.Pipeline
}

func GenerateHandle(l int) string {

	b := make([]rune, l)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func ParseHexID(param string) (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex(param)
}
func RespondWithError(c *fiber.Ctx, code int, msg string) error {
	return c.Status(code).JSON(fiber.Map{"error": msg})
}

func BuildImgURL(imageID string) string {
	return fmt.Sprintf("%s/%s", baseImgURL, imageID)
}

func VerifyOwnership(c *fiber.Ctx, ownerID primitive.ObjectID) (bool, error) {
	user := c.Locals("jwt").(*jwt.Token)

	claims := user.Claims.(jwt.MapClaims)
	userID, err := ParseHexID(claims["id"].(string))
	if err != nil {
		return false, err
	}

	fmt.Printf("Claims: %s\n", claims)
	fmt.Printf("Owner ID: %s\n", ownerID)
	fmt.Printf("User ID: %s\n", userID)

	if userID != ownerID {
		return false, nil
	}

	return true, nil

}

func GetUserID(c *fiber.Ctx) (primitive.ObjectID, error) {
	user := c.Locals("jwt").(*jwt.Token)

	claims := user.Claims.(jwt.MapClaims)
	userID, err := ParseHexID(claims["id"].(string))
	if err != nil {
		return primitive.ObjectID{}, err
	}

	return userID, nil
}

func NewPipeline() *PipelineBuilder {
	return &PipelineBuilder{stages: mongo.Pipeline{}}
}

func (pb *PipelineBuilder) Sort(field string, order int) *PipelineBuilder {
	pb.stages = append(pb.stages, bson.D{{"$sort", bson.D{{field, order}}}})
	return pb
}

func (pb *PipelineBuilder) Match(field bson.D) *PipelineBuilder {
	pb.stages = append(pb.stages, bson.D{{"$match", field}})
	return pb
}

func (pb *PipelineBuilder) Skip(n int64) *PipelineBuilder {
	pb.stages = append(pb.stages, bson.D{{"$skip", n}})
	return pb
}

func (pb *PipelineBuilder) Limit(n int64) *PipelineBuilder {
	pb.stages = append(pb.stages, bson.D{{"$limit", n}})
	return pb
}

func (pb *PipelineBuilder) Lookup(from, localField, foreignField, as string) *PipelineBuilder {
	pb.stages = append(pb.stages, bson.D{{"$lookup", bson.D{
		{"from", from},
		{"localField", localField},
		{"foreignField", foreignField},
		{"as", as},
	}}})

	return pb
}

func (pb *PipelineBuilder) Unwind(path string, preserve bool) *PipelineBuilder {
	pb.stages = append(pb.stages, bson.D{{"$unwind", bson.D{
		{"path", path},
		{"preserveNullAndEmptyArrays", preserve},
	}}})

	return pb
}

func (pb *PipelineBuilder) Project(fields bson.D) *PipelineBuilder {
	pb.stages = append(pb.stages, bson.D{{"$project", fields}})
	return pb
}

func (pb *PipelineBuilder) Build() mongo.Pipeline {
	return pb.stages
}
