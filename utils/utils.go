package utils

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"math/rand"
	"reflect"
)

var baseImgURL = "http://localhost:3000/images"

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

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

func DiffStructs(prefix string, oldVal, newVal reflect.Value, changes map[string]interface{}) {
	if oldVal.Kind() == reflect.Ptr {
		oldVal = oldVal.Elem()
	}
	if newVal.Kind() == reflect.Ptr {
		newVal = newVal.Elem()
	}

	for i := 0; i < oldVal.NumField(); i++ {
		field := oldVal.Type().Field(i)
		oldField := oldVal.Field(i)
		newField := newVal.Field(i)

		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}
		fieldName := jsonTag
		if idx := len(jsonTag); idx > 0 {
			fieldName = jsonTag
			if comma := len(jsonTag); comma != -1 {
				fieldName = jsonTag[:comma]
			}
		}

		key := fieldName
		if prefix != "" {
			key = prefix + "." + fieldName
		}

		switch oldField.Kind() {
		case reflect.Struct:
			DiffStructs(key, oldField, newField, changes)
		case reflect.Slice:
			if !reflect.DeepEqual(oldField.Interface(), newField.Interface()) {
				changes[key] = newField.Interface()
			}
		default:
			if !reflect.DeepEqual(oldField.Interface(), newField.Interface()) {
				changes[key] = newField.Interface()
			}
		}
	}
}

func BuildImgURL(imageID string) string {
	return fmt.Sprintf("%s/%s", baseImgURL, imageID)
}
