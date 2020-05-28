package v1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"set-flags/models"
	"set-flags/pkg/e"
	"set-flags/pkg/setting"
)

// list all the flags
func ListFlags(c *gin.Context) {
	code := e.INVALID_PARAMS
	data := models.GetAllFlags()
	code = e.SUCCESS
	c.JSON(http.StatusOK, gin.H{
		"code": code,
		"msg": e.GetMsg(code),
		"data": data,
	})
}

// create a flag
func CreateFlag(c *gin.Context) {

	var flag map[string]interface{}

	if c.ShouldBind(&flag) == nil {
		models.CreateFlag(flag)
	}
	c.JSON(http.StatusCreated, gin.H{
		"code": 1,
		"msg":  "created flag",
		"data": make(map[string]interface{}),
	})
}

// Update an existing flag
func UpdateFlag(c *gin.Context) {
	flagId := c.Param("id")

	_, err := uuid.FromString(flagId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"info": fmt.Sprintf("flagId: %s is not a valid UUID.", flagId),
		})
		return
	}

	op := c.Param("op")

	if op != "yes" && op != "no" && op != "done" {
		c.JSON(http.StatusBadRequest, gin.H{
			"info": fmt.Sprintf("op: %s is invalid.", op),
		})
		return
	}

	if !models.FLagExists(flagId) {
		c.JSON(http.StatusNotFound, gin.H{
			"info": "Flag not found.",
		})
		return
	}

	flag := models.FindFlagByID(flagId)

	if flag.Status != "done" {
		c.JSON(http.StatusBadRequest, gin.H{
			"info": "not yet upload the evidence.",
		})
		return
	}

}

// list all flags of the user
func FindFlagsByUserID(c *gin.Context) {
	userId := c.Param("id")

	flags := models.FindFlagsByUserID(userId)
	c.PureJSON(http.StatusOK, flags)
}

// upload evidence
func UploadEvidence(c *gin.Context) {
	code := e.INVALID_PARAMS
	flagId := c.Query("flag_id")
	attachmentId := c.Param("attachment_id")

	fmt.Sprintf("attachmentId: %s, flagId: %s", attachmentId, flagId)

	type_ := c.Query("type")

	if type_ != "image" && type_ != "audio" && type_ != "video" && type_ != "document" {
		c.JSON(http.StatusBadRequest, gin.H{
			"info": fmt.Sprintf("type: %s is invalid.", type_),
		})
		return
	}

	if !models.FLagExists(flagId) {
		c.JSON(http.StatusNotFound, gin.H{
			"info": "not found specific flag.",
		})
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"info": err,
		})
		return
	}

	log.Println(fileHeader.Filename)

	client := &http.Client{}
	accessToken := ""

	viewUrl, err := UploadAttachment(client, fileHeader, accessToken)
	if err != nil {
		code = e.ERROR_UPLOAD_ATTACHMENT
		c.JSON(http.StatusOK, gin.H{
			"code": code,
			"msg": e.GetMsg(code),
			"data": make(map[string]interface{}),
		})
		return
	}

	attachmentID, _ := uuid.FromString(attachmentId)
	flagID, _ := uuid.FromString(flagId)
	models.CreateEvidence(attachmentID, flagID, type_, viewUrl)

	// update flag status to `done`
	models.UpdateFlagStatus(flagId, "done")

	c.JSON(http.StatusOK, gin.H{
		"info": fmt.Sprintf("'%s' uploaded!", fileHeader.Filename),
	})
}

// list all the evidences since yesterday
func ListEvidences(c *gin.Context) {
	flagId := c.Param("flag_id")
	flagID, _ := uuid.FromString(flagId)
	data := models.FindEvidencesByFlag(flagID)

	code := 200
	c.JSON(http.StatusOK, gin.H{
		"code": code,
		"msg": e.GetMsg(code),
		"data": data,
	})
}

// https://developers.mixin.one/api/l-messages/create-attachment/
func UploadAttachment(client *http.Client, fileHeader *multipart.FileHeader, accessToken string) (string, error) {

	f, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	size := fileHeader.Size

	buffer := make([]byte, size)
	_, err = f.Read(buffer)

	if err != nil {
		return "", err
	}


	url := fmt.Sprintf("%s/attachments", setting.MixinAPIDomain)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(buffer))

	req.Header.Add("Content-Type", http.DetectContentType(buffer))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	resp, err := client.Do(req)

	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var authResp map[string]map[string]interface{}

	data, _ := ioutil.ReadAll(resp.Body)

	_ = json.Unmarshal(data, &authResp)

	viewUrl, _ := authResp["data"]["view_url"].(string)

	return viewUrl, nil
}