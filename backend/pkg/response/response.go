package response

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/iamtbay/tyr-fintech/pkg/apperrors"
)

func HandleValidationError(c *gin.Context, err error) {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		errs := make(map[string]string)
		for _, fe := range ve {
			errs[fe.Field()] = getValidationErrorMsg(fe)
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"errors":  errs,
		})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{
		"success": false,
		"error":   "Invalid request",
	})
}

func getValidationErrorMsg(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fe.Field() + " is required"
	case "email":
		return fe.Field() + "must be a valid email address"
	case "min":
		return fe.Field() + fmt.Sprintf("must be at least %v characters long", fe.Param())
	case "max":
		return fmt.Sprintf("must be at most %v characters long", fe.Param())
	default:
		return "invalid value"
	}
}

func HandleError(c *gin.Context, err error) {
	if appErr, ok := err.(*apperrors.AppError); ok {
		c.JSON(appErr.Code, gin.H{
			"success": false,
			"error":   appErr.Msg,
		})
		return
	}

	c.JSON(http.StatusInternalServerError, gin.H{
		"success": false,
		"error":   "unexpected error occured",
	})
}

func Success(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{
		"success": true,
		"data":    data,
	})
}

func Error(c *gin.Context, status int, errMessage string) {
	c.JSON(status, gin.H{
		"success": false,
		"error":   errMessage,
	})
}
