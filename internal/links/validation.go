package links

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
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
	fields := make(map[string]string)
	if err := binding.Validator.ValidateStruct(payload); err != nil {
		var violations validator.ValidationErrors
		if !errors.As(err, &violations) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate request"})
			return false
		}

		typ := reflect.TypeOf(payload).Elem()
		for _, violation := range violations {
			field, _ := typ.FieldByName(violation.StructField())
			name := strings.Split(field.Tag.Get("json"), ",")[0]
			fields[name] = violation.Error()
		}
	}
	value := reflect.ValueOf(payload).Elem()
	if !validRedirectURL(value.FieldByName("OriginalURL").String()) {
		fields["original_url"] = "must be an absolute http or https URL with a valid host"
	}
	if name := value.FieldByName("ShortName").String(); name != "" && !shortNamePattern.MatchString(name) {
		fields["short_name"] = "must contain 3 to 32 ASCII letters, digits, underscores or hyphens"
	}
	if len(fields) > 0 {
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

var hostLabelPattern = regexp.MustCompile(`^[A-Za-z0-9](?:[A-Za-z0-9-]*[A-Za-z0-9])?$`)

var shortNamePattern = regexp.MustCompile(`^[A-Za-z0-9_-]{3,32}$`)

func validRedirectURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Opaque != "" || u.Hostname() == "" {
		return false
	}
	host := u.Hostname()
	if net.ParseIP(host) == nil {
		if strings.HasPrefix(u.Host, "[") {
			return false
		}
		for _, label := range strings.Split(strings.TrimSuffix(host, "."), ".") {
			if len(label) > 63 || !hostLabelPattern.MatchString(label) {
				return false
			}
		}
		if len(host) > 253 {
			return false
		}
	}
	if port := u.Port(); port != "" {
		n, err := strconv.Atoi(port)
		if err != nil || n < 1 || n > 65535 {
			return false
		}
	}
	return true
}
