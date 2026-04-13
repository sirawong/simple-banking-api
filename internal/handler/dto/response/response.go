package response

import (
	"errors"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"

	pkgerrs "github.com/sirawong/simple-banking-api/pkg/errs"
)

type ResponseType int

const (
	TypeDefault ResponseType = iota
	TypeCreated
)

var trans ut.Translator

func HandleResponse(c *gin.Context, data any, err error) {
	handleWithType(c, data, err, TypeDefault)
}

func HandleCreatedResponse(c *gin.Context, data any, err error) {
	handleWithType(c, data, err, TypeCreated)
}

func handleWithType(c *gin.Context, data any, err error, responseType ResponseType) {
	if err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			errorDetails := make(map[string]string)
			for _, e := range validationErrs {
				errorDetails[e.Field()] = e.Translate(trans)
			}
			appErr := pkgerrs.ErrBadRequest.WithDetails(errorDetails)
			c.JSON(appErr.HttpStatusCode, appErr)
			return
		}

		var appErr *pkgerrs.AppError
		if errors.As(err, &appErr) {
			c.JSON(appErr.HttpStatusCode, appErr)
		} else {
			wrapped := pkgerrs.ErrInternal.WithError(err)
			c.JSON(http.StatusInternalServerError, wrapped)
		}
		return
	}

	statusCode := successStatusCode(responseType, data != nil)
	if data != nil {
		c.JSON(statusCode, data)
	} else {
		c.Status(statusCode)
	}
}

func successStatusCode(responseType ResponseType, hasData bool) int {
	if responseType == TypeCreated {
		return http.StatusCreated
	}
	if hasData {
		return http.StatusOK
	}
	return http.StatusNoContent
}

func init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		enLocale := en.New()
		uni := ut.New(enLocale, enLocale)
		trans, _ = uni.GetTranslator("en")
		_ = enTranslations.RegisterDefaultTranslations(v, trans)

		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})

		_ = v.RegisterTranslation("required", trans, func(ut ut.Translator) error {
			return ut.Add("required", "{0} is a required field", true)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("required", fe.Field())
			return t
		})

		_ = v.RegisterTranslation("gt", trans, func(ut ut.Translator) error {
			return ut.Add("gt", "{0} must be greater than {1}", true)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("gt", fe.Field(), fe.Param())
			return t
		})
	}
}
