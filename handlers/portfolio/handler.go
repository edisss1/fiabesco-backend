package portfolio

import (
	"context"
	"github.com/edisss1/fiabesco-backend/db"
	"github.com/edisss1/fiabesco-backend/types"
	"github.com/edisss1/fiabesco-backend/utils"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var collection *mongo.Collection

func CreatePortfolio(c *fiber.Ctx) error {

	var body types.Portfolio

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondWithError(c, 400, "Invalid request body")
	}

	collection = db.Database.Collection("portfolios")

	_, err := collection.InsertOne(context.Background(), body)
	if err != nil {
		return utils.RespondWithError(c, 500, "Failed to create portfolio")
	}

	return c.Status(201).JSON(fiber.Map{"msg": "Portfolio created successfully"})

}

func GetPortfolio(c *fiber.Ctx) error {
	id := c.Params("userID")
	userID, err := utils.ParseHexID(id)
	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid user ID")
	}

	filter := bson.M{"userID": userID}
	collection = db.Database.Collection("portfolios")

	var portfolio types.Portfolio

	err = collection.FindOne(context.Background(), filter).Decode(&portfolio)
	if err != nil {
		return utils.RespondWithError(c, 404, "Portfolio not found")
	}

	return c.Status(200).JSON(portfolio)

}

func UpdatePortfolio(c *fiber.Ctx) error {
	//id := c.Params("userID")
	//userID, err := utils.ParseHexID(id)
	//if err != nil {
	//	return utils.RespondWithError(c, 400, "Invalid user ID")
	//}
	return nil
}
