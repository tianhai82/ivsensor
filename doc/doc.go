package doc

import (
	"context"
	"fmt"
	"net/http"

	"cloud.google.com/go/storage"
	"github.com/gin-gonic/gin"
	"github.com/plandem/xlsx"
	"github.com/tianhai82/ivsensor/crawler"
	"github.com/tianhai82/ivsensor/firebase"
	"github.com/tianhai82/ivsensor/model"
)

func GenDoc(c *gin.Context) {
	date := c.Param("date")

	docIter := firebase.FirestoreClient.Collection("record").Where("Date", "==", date).Documents(context.Background())
	docs, err := docIter.GetAll()
	if err != nil {
		fmt.Println("fail to retrieve records from firestore", err)
		c.AbortWithError(500, err)
		return
	}

	excel := xlsx.New()
	sheet := excel.AddSheet("options")
	crawler.WriteHeader(sheet)
	row := 1
	for _, doc := range docs {
		var rec model.OptionRecord
		err = doc.DataTo(&rec)
		if err != nil {
			fmt.Println(err)
			continue
		}
		crawler.WriteRecord(sheet, row, rec)
		row++
	}

	filename := fmt.Sprintf("%s.xlsx", date)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	excel.SaveAs(c.Writer)
	excel.Close()
}

func HandleDownload(c *gin.Context) {
	filename := c.Param("filename")
	reader, err := retrieveFile(filename)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("file not found"))
		return
	}
	defer reader.Close()

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", reader.Attrs.ContentType)
	c.DataFromReader(200, reader.Attrs.Size, reader.Attrs.ContentType, reader, nil)
}

func retrieveFile(filename string) (*storage.Reader, error) {
	bucket, err := firebase.StorageClient.DefaultBucket()
	if err != nil {
		fmt.Println("fail to get bucket", err)
		return nil, err
	}
	return bucket.Object(filename).NewReader(context.Background())
}
