package utils

import (
	"fmt"
	"myproject/api/database"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func SaveImage(c *fiber.Ctx, formName string, existingPath string, model string, country string, province string, id uint) (string, error) {
	file, err := c.FormFile(formName) // formName is the name of the file in the multipart form
	if err == nil {
		// has an image
		ext := filepath.Ext(file.Filename)

		name := strings.Map(CheckChars, file.Filename)
		country = strings.Map(CheckChars, country)
		province = strings.Map(CheckChars, province)

		key := fmt.Sprintf("%s/%s/business/%d/%s/%s", country, province, id, model, name)

		if model == "bin" {
			key = fmt.Sprintf("%s/%s/bin/%d/%s", country, province, id, name)
		}

		if strings.HasSuffix(key, ext) {

		} else {
			key = fmt.Sprintf("%s%s", key, ext)
		}

		key = strings.Map(CheckChars, key)

		if os.Getenv("USE_DOCKER") == "true" {
			key = fmt.Sprintf("docker/%s", key)
			fmt.Println("save image to spaces/docker", key)
		} else if os.Getenv("TEST_MODE") == "true" || database.GetParam("DEV_MODE") == "true" {
			key = fmt.Sprintf("test/%s", key)
			fmt.Println("save image to spaces/test", key)
		}

		var result string

		if result, err = saveFileSystemImage(c, strings.Map(CheckChars, formName), existingPath, model); err != nil {
			fmt.Println(err)
			c.Status(503).SendString(err.Error())
			return "", err
		}
		return result, nil
	}
	return "", err
}

func copyLocalImage(existingPath string, id uint) (string, error) {

	dir, file := path.Split(existingPath)

	key := fmt.Sprintf("%s/%s ", dir, file)
	key = strings.Map(CheckChars, key)

	return key, nil
}

func CopyImage(existingPath string, id uint) (string, error) {

	return copyLocalImage(existingPath, id)
}

func DeleteImage(path string) {
	if os.Getenv("USE_DOCKER") == "true" {
		fmt.Println("USE_DOCKER - NOT deleting ", path)
		return
	} else if os.Getenv("TEST_MODE") == "true" || database.GetParam("DEV_MODE") == "true" {
		fmt.Println("DEV_MODE or TEST_MODE - NOT deleting ", path)
		return
	}

	deleteFileSystemImage(path)
}

func SaveAsset(c *fiber.Ctx, formName string, existingPath string, country string, province string, id uint) (string, error) {
	return SaveImage(c, formName, existingPath, "asset", country, province, id)
}
