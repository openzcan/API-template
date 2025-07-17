package utils

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

//////// filesystem images

func mkdir(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		mdErr := os.MkdirAll(path, 0755)
		if mdErr != nil {
			fmt.Println("Error making directory", mdErr)
			return false
		}
	}
	return true
}

func deleteFileSystemImage(path string) {
	if path == "" {
		return
	}
	path = fmt.Sprintf(".%s", path)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		//fmt.Println(path, "does not exist")
		return
	}
	//fmt.Println(path, "removed")
	os.Remove(path)
}

func saveFile(c *fiber.Ctx, file *multipart.FileHeader, path string) error {
	savedUmask := syscall.Umask(0222)
	if err := c.SaveFile(file, path); err != nil {
		return err
	}
	//text := fmt.Sprintf("old umask %d %o", savedUmask, savedUmask)
	//fmt.Println("old umask", text)
	if err := os.Chmod(path, 0644); err != nil {
		fmt.Println(err)
	}
	_ = syscall.Umask(savedUmask) // Return the umask to the original
	return nil
}

func saveFileSystemImage(c *fiber.Ctx, formName string, existingPath string, model string) (string, error) {
	file, err := c.FormFile(formName) // formName is the name of the file in the multipart form
	if err == nil {
		// has an image
		ext := filepath.Ext(file.Filename)
		// make the user file dir
		rand, _ := uuid.NewRandom()
		fname := fmt.Sprintf("%s%s", rand, ext) // ext includes the leading '.'

		path := fmt.Sprintf("user_files/%s/%s/%s", model,
			fname[0:2], fname[2:4])

		if mkdir(path) {

			fullpath := fmt.Sprintf("%s/%s", path, fname)

			if err := saveFile(c, file, fullpath); err != nil {
				return "", err
			}
			// delete previous image if not a default image
			if "" != existingPath && strings.Contains(existingPath, "user_files") {
				//fmt.Println("######### delete existing image", existingPath)
				deleteFileSystemImage(existingPath)
			}

			// File.delete(image.file.path)
			return fmt.Sprintf("/%s", fullpath), nil
		}
	}
	return "", err
}
