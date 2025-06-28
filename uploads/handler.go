package uploads

import (
	"bytes"
	"github.com/edisss1/fiabesco-backend/db"
	"github.com/edisss1/fiabesco-backend/utils"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"io"
	"net/http"
)

func UploadFile(c *fiber.Ctx, field string, bucket *gridfs.Bucket, allowMultiple bool) ([]primitive.ObjectID, error) {
	var uploadedIDs []primitive.ObjectID

	if allowMultiple {
		form, err := c.MultipartForm()
		if err != nil {
			return nil, err
		}
		files := form.File[field]
		for _, fh := range files {
			file, err := fh.Open()
			if err != nil {
				return nil, err
			}
			defer file.Close()

			uploadStream, err := bucket.OpenUploadStream(fh.Filename)
			if err != nil {
				return nil, err
			}
			defer uploadStream.Close()
			io.Copy(uploadStream, file)

			uploadedIDs = append(uploadedIDs, uploadStream.FileID.(primitive.ObjectID))
		}
	} else {
		fh, err := c.FormFile(field)
		if err != nil {
			return nil, err
		}
		file, err := fh.Open()
		if err != nil {
			return nil, err
		}
		defer file.Close()

		uploadStream, err := bucket.OpenUploadStream(fh.Filename)
		if err != nil {
			return nil, err
		}
		defer uploadStream.Close()
		io.Copy(uploadStream, file)
		uploadedIDs = append(uploadedIDs, uploadStream.FileID.(primitive.ObjectID))

	}
	return uploadedIDs, nil
}

func ServeImage(c *fiber.Ctx) error {
	id := c.Params("imageID")
	imageID, err := utils.ParseHexID(id)
	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid image ID")
	}

	bucket, err := gridfs.NewBucket(db.Database)
	if err != nil {
		return utils.RespondWithError(c, 500, "Failed to create bucket")
	}

	var buf bytes.Buffer
	_, err = bucket.DownloadToStream(imageID, &buf)
	if err != nil {
		return utils.RespondWithError(c, 500, "Failed to download image")
	}
	contentType := http.DetectContentType(buf.Bytes())
	c.Set("Content-Type", contentType)
	c.Set("Content-Disposition", "inline")

	return c.Send(buf.Bytes())

}
