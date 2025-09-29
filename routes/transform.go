package routes

import (
	"fmt"
	"lazy-lagoon/pkg/httphelper"
	"lazy-lagoon/pkg/types"
	"lazy-lagoon/transformcsv"
	"lazy-lagoon/transformjson"
	"net/http"

	"github.com/gin-gonic/gin"
)

/*
Transform the file - used for preview and snapshot mutations
*/
func Transform(c *gin.Context) {
	/*
		Request body
	*/
	var requestData types.RequestBodyTransform
	var err error

	err = bindAndValidate(c, &requestData)
	if err != nil {
		sendError(c, http.StatusBadRequest, err, nil)
		return
	}

	input := requestData.Input
	dataType := input.DataType
	rules := requestData.Rules
	output := requestData.Output
	webhook := requestData.Webhook

	/*
		Split the process into different data types
	*/
	var transformErr *types.TransformError
	switch dataType {
	case "CSV", "SQL":
		transformErr = transformcsv.ExecuteTransform(input, rules, output)
		if transformErr != nil {
			sendTransformError(c, http.StatusInternalServerError, transformErr, webhook)
			return
		}
	case "JSON":
		transformErr = transformjson.ExecuteTransform(input, rules, output)
		if transformErr != nil {
			sendTransformError(c, http.StatusInternalServerError, transformErr, webhook)
			return
		}
	case "JSONL":
		transformErr = transformjson.ExecuteTransformJsonl(input, rules, output)
		if transformErr != nil {
			sendTransformError(c, http.StatusInternalServerError, transformErr, webhook)
			return
		}
	default:
		sendError(c, http.StatusBadRequest, fmt.Errorf("data type %s not found", dataType), webhook)
		return
	}

	// Request is async
	if webhook != nil {
		/*
			Sending the webhook
		*/
		err = httphelper.SendPostRequest(webhook.Payload, webhook.Url, webhook.ResponseToken)
		if err != nil {
			sendError(c, http.StatusInternalServerError, err, webhook)
			return
		}
	}

	c.JSON(http.StatusOK, "Completed transformation")
	return
}
