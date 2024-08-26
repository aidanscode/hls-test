package main

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
)

var allowedUploadTypes = map[string]string{
	"audio/mpeg": "mp3",
	"video/mp4": "mp4",
	"video/webm": "webm",
	"audio/wav": "wav",
}

type Renderer struct {
	templates *template.Template
}

func (r *Renderer) Render(w io.Writer, name string, data interface{}, con echo.Context) error {
	return r.templates.ExecuteTemplate(w, name, data)
}

func NewRenderer() *Renderer {
	return &Renderer{
		templates: template.Must(template.ParseGlob("views/*.html")),
	}
}

type UploadReceivedData struct {
	Errors []string
}

func NewUploadReceivedData(error string) UploadReceivedData {
	return UploadReceivedData{Errors: []string{error}}
}

func main() {
	e := echo.New()
	e.Renderer = NewRenderer()

	e.GET("/", func(c echo.Context) error {
		return c.HTML(200, "<a href=\"/upload\">Upload new media</a>")
	})

	e.GET("/upload", func(c echo.Context) error {
		return c.Render(200, "upload", nil)
	})

	e.POST("/upload", func(c echo.Context) error {
		name := c.FormValue("name")
		errors := make([]string, 0)
		if name == "" {
			errors = append(errors, "Missing name!")
		}

		file, err := c.FormFile("media")
		if file == nil || err != nil {
			errors = append(errors, "Missing or invalid file")
		}

		if len(errors) > 0 {
			return c.Render(http.StatusUnprocessableEntity, "upload-received", UploadReceivedData{Errors: errors})
		}

		src, err := file.Open()
		if err != nil {
			return c.Render(http.StatusUnprocessableEntity, "upload-received", NewUploadReceivedData("Failed to parse file"))
		}
		defer src.Close()

		extension, err := validateMimeType(src)
		if err != nil {
			return c.Render(http.StatusUnprocessableEntity, "upload-received", NewUploadReceivedData(err.Error()))
		}

		path := fmt.Sprintf("media/%v.%s", time.Now().UnixMilli(), extension)
		newFile, err := os.Create(path)
		if err != nil {
			return c.Render(http.StatusUnprocessableEntity, "upload-received", NewUploadReceivedData(err.Error()))
		}
		defer newFile.Close()

		if _, err := io.Copy(newFile, src); err != nil {
			return c.Render(http.StatusUnprocessableEntity, "upload-received", NewUploadReceivedData(err.Error()))
		}

		return c.String(200, "Saved in " + path)
	})

	e.Logger.Fatal(e.Start(":8000"))
}

func validateMimeType(f multipart.File) (string, error) {
	contents, err := io.ReadAll(f)
	defer f.Seek(0, io.SeekStart)
	if err != nil {
		return "", errors.New("invalid image uploaded")
	}

	mimeType := http.DetectContentType(contents)
	ext, ok := allowedUploadTypes[mimeType]
	if !ok {
		return "", errors.New("invalid file type. Must be one of: mp3, mp4, webm, or wav")
	}
	return ext, nil
}
