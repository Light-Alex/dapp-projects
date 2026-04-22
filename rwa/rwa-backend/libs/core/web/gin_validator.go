package web

import (
	"context"
	"reflect"
	"strings"

	"github.com/AnchoredLabs/rwa-backend/libs/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
	zhTranslations "github.com/go-playground/validator/v10/translations/zh"
)

const (
	evmAddress = "checkEvmAddress"
)

var validateEvmAddress validator.Func = func(fl validator.FieldLevel) bool {
	address, ok := fl.Field().Interface().(string)
	if ok {
		address = strings.ToLower(address)
		return common.IsHexAddress(address)
	}
	return true
}

func BindGinCusValidator(_ context.Context, defaultLocale string) (ut.Translator, error) {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return nil, errors.Errorf("failed to initialize translator")
	}
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	zhT := zh.New()
	enT := en.New()
	uni := ut.New(enT, zhT, enT)
	trans, ok := uni.GetTranslator(defaultLocale)
	if !ok {
		return nil, errors.Errorf("uni.FindTranslator(%s) failed", defaultLocale)
	}
	var err error
	switch defaultLocale {
	case "en":
		err = enTranslations.RegisterDefaultTranslations(v, trans)
	case "zh":
		err = zhTranslations.RegisterDefaultTranslations(v, trans)
	default:
		err = enTranslations.RegisterDefaultTranslations(v, trans)
	}
	if err != nil {
		return nil, errors.Annotatef(err, "register validator %s error", defaultLocale)
	}
	err = v.RegisterValidation(evmAddress, validateEvmAddress)
	if err != nil {
		return nil, errors.Errorf("register validator %s error", evmAddress)
	}
	// register translation for evmAddress
	if err = v.RegisterTranslation(
		evmAddress,
		trans,
		RegisterTranslator(evmAddress, "{0} not a valid evm address"),
		Translate,
	); err != nil {
		return nil, err
	}
	return trans, err
}

func RemoveTopStruct(fields map[string]string) map[string]string {
	res := map[string]string{}
	for field, err := range fields {
		res[field[strings.Index(field, ".")+1:]] = err
	}
	return res
}

// RegisterTranslator custom register translator
func RegisterTranslator(tag string, msg string) validator.RegisterTranslationsFunc {
	return func(trans ut.Translator) error {
		if err := trans.Add(tag, msg, false); err != nil {
			return err
		}
		return nil
	}
}

// Translate custom translate function
func Translate(trans ut.Translator, fe validator.FieldError) string {
	msg, _ := trans.T(fe.Tag(), fe.Field())
	return msg
}
