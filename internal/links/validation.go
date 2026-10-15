package links

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// Decode the whole document before validation so malformed JSON always gets 400.
func bindLinkRequest(c *gin.Context, payload any) bool {
	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return false
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return false
	}
	if err := binding.Validator.ValidateStruct(payload); err != nil {
		var violations validator.ValidationErrors
		if !errors.As(err, &violations) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate request"})
			return false
		}
		fields := make(map[string]string, len(violations))
		typ := reflect.TypeOf(payload).Elem()
		for _, violation := range violations {
			field, _ := typ.FieldByName(violation.StructField())
			name := strings.Split(field.Tag.Get("json"), ",")[0]
			fields[name] = violation.Error()
		}
		c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": fields})
		return false
	}
	return true
}

func writeLinkError(c *gin.Context, err error, message string) {
	if errors.Is(err, ErrLinkNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": ErrLinkNotFound.Error()})
		return
	}
	if errors.Is(err, ErrShortNameInUse) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"errors": gin.H{"short_name": ErrShortNameInUse.Error()},
		})
		return
	}
	log.Printf("%s: %v", message, err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": message})
}
