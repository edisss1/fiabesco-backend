package portfolio

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/edisss1/fiabesco-backend/db"
	"github.com/edisss1/fiabesco-backend/types"
	"github.com/edisss1/fiabesco-backend/uploads"
	"github.com/edisss1/fiabesco-backend/utils"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"io"
)

var collection *mongo.Collection

func CreatePortfolio(c *fiber.Ctx) error {
	userID := c.Params("userID")

	portfolioJSON := c.FormValue("portfolio")
	var portfolio types.Portfolio
	if err := json.Unmarshal([]byte(portfolioJSON), &portfolio); err != nil {
		return utils.RespondWithError(c, 400, "Invalid portfolio data")
	}

	var existingPortfolio types.Portfolio
	err := collection.FindOne(context.Background(), bson.M{"userID": userID}).Decode(&existingPortfolio)

	if err == nil {
		return utils.RespondWithError(c, 400, "Portfolio already exists")
	}

	bucket, err := gridfs.NewBucket(db.Database)
	if err != nil {
		return utils.RespondWithError(c, 500, "Failed to create bucket")
	}

	for i := range portfolio.Projects {
		fieldName := fmt.Sprintf("project-img-%d", i)
		ids, err := uploads.UploadFile(c, fieldName, bucket, false)
		if err != nil {
			return utils.RespondWithError(c, 500, "Failed to upload file"+err.Error())
		}

		if len(ids) > 0 {
			portfolio.Projects[i].Img = ids[0].Hex()
		}
	}

	portfolio.UserID = userID
	collection = db.Database.Collection("portfolios")
	_, err = collection.InsertOne(context.Background(), portfolio)
	if err != nil {
		return utils.RespondWithError(c, 500, "Failed to create portfolio "+err.Error())
	}

	return c.Status(201).JSON(fiber.Map{"msg": "Portfolio created successfully"})

}

func GetPortfolio(c *fiber.Ctx) error {
	userID := c.Params("userID")

	var portfolio types.Portfolio
	collection = db.Database.Collection("portfolios")
	filter := bson.M{"userID": userID}
	err := collection.FindOne(context.Background(), filter).Decode(&portfolio)
	if err != nil {

		return utils.RespondWithError(c, 500, "Failed to get portfolio "+err.Error())
	}

	bucket, err := gridfs.NewBucket(db.Database)
	if err != nil {
		return utils.RespondWithError(c, 500, "Failed to create bucket")
	}

	for i, project := range portfolio.Projects {
		fileID, err := utils.ParseHexID(project.Img)
		if err != nil {
			continue
		}
		var buf bytes.Buffer
		downloadStream, err := bucket.OpenDownloadStream(fileID)
		if err != nil {
			continue
		}
		io.Copy(&buf, downloadStream)
		downloadStream.Close()
		base64Img := base64.StdEncoding.EncodeToString(buf.Bytes())
		portfolio.Projects[i].Img = "data:image/jpeg;base64," + base64Img
	}

	var user types.User
	collection = db.Database.Collection("users")

	parsedUserID, err := utils.ParseHexID(userID)
	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid user ID")
	}
	filter = bson.M{"_id": parsedUserID}

	err = collection.FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		return utils.RespondWithError(c, 500, "Failed to get user "+err.Error())
	}
	portfolio.UserName = user.FirstName + " " + user.LastName

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
