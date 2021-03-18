package doc

import (
	"context"
	"fmt"
	"net/http"

	"cloud.google.com/go/storage"
	"github.com/gin-gonic/gin"
	"github.com/tianhai82/ivsensor/firebase"
)

func HandleDownload(c *gin.Context) {
	filename := c.Param("filename")
	reader, err := retrieveFile(filename)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("file not found"))
		return
	}
	defer reader.Close()

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", reader.ContentType())
	c.DataFromReader(200, reader.Size(), reader.ContentType(), reader, nil)
}

func retrieveFile(filename string) (*storage.Reader, error) {
	bucket, err := firebase.StorageClient.DefaultBucket()
	if err != nil {
		fmt.Println("fail to get bucket", err)
		return nil, err
	}
	return bucket.Object(filename).NewReader(context.Background())
}
