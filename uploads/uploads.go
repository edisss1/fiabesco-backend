package uploads

import (
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"io"
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
