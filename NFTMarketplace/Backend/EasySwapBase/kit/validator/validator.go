package validator

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	ent "github.com/go-playground/validator/v10/translations/en"
)

const (
	bankcardRegexString    = "^[0-9]{15,19}$"
	corpaccountRegexString = "^[0-9]{9,25}$"
	idcardRegexString      = "^[0-9]{17}[0-9X]$"
	usccRegexString        = "^[A-Z0-9]{18}$"
)

var (
	// ErrUnexpected 校验意外错误
	ErrUnexpected = errors.New("err Unexpected")

	ti ut.Translator
	vi *validator.Validate

	bankcardRegex    = regexp.MustCompile(bankcardRegexString)
	corpaccountRegex = regexp.MustCompile(corpaccountRegexString)
	idcardRegex      = regexp.MustCompile(idcardRegexString)
	usccRegex        = regexp.MustCompile(usccRegexString)

	httpMethodMap = map[string]struct{}{
		"GET":     {},
		"POST":    {},
		"PUT":     {},
		"DELETE":  {},
		"PATCH":   {},
		"OPTIONS": {},
	}
	defaultTags = []string{
		"required_if",
		"required_unless",
		"required_with",
		"required_with_all",
		"required_without",
		"required_without_all",
		"excluded_with",
		"excluded_with_all",
		"excluded_without",
		"excluded_without_all",
		"isdefault",
		"fieldcontains",
		"fieldexcludes",
		"boolean",
		"e164",
		"urn_rfc2141",
		"file",
		"base64url",
		"startsnotwith",
		"endsnotwith",
		"eth_addr",
		"btc_addr",
		"btc_addr_bech32",
		"uuid_rfc4122",
		"uuid3_rfc4122",
		"uuid4_rfc4122",
		"uuid5_rfc4122",
		"hostname",
		"hostname_rfc1123",
		"fqdn",
		"unique",
		"html",
		"html_encoded",
		"url_encoded",
		"dir",
		"jwt",
		"hostname_port",
		"timezone",
		"iso3166_1_alpha2",
		"iso3166_1_alpha3",
		"iso3166_1_alpha_numeric",
		"iso3166_2",
		"iso4217",
		"iso4217_numeric",
		"bcp47_language_tag",
		"postcode_iso3166_alpha2",
		"postcode_iso3166_alpha2_field",
		"bic",
	}
)

type ValidateErrors []string

// Error 返回验证错误字符串并实现error接口，默认只返回第一个字段的验证错误
func (ves ValidateErrors) Error() string {
	// 如果ValidateErrors长度大于0
	if len(ves) > 0 {
		// 返回第一个错误消息
		return ves[0]
	}

	// 否则返回意外错误消息
	return ErrUnexpected.Error()
}

// ParseErr 解析验证错误具体内容
func ParseErr(err error) string {
	// 判断错误是否可以被断言为ValidateErrors类型
	ves, ok := err.(ValidateErrors)
	if ok && len(ves) > 0 {
		// 如果错误可以被断言为ValidateErrors类型且其长度大于0，则将错误拼接为字符串并返回
		return strings.Join(ves, ",")
	}

	// 如果错误不能被断言为ValidateErrors类型或其长度为0，则返回错误的默认描述
	return err.Error()
}

func init() {
	var err error
	// 创建一个新的验证器实例
	vi = validator.New()
	// 设置标签名为 "validate"
	vi.SetTagName("validate")
	// 注册标签名获取函数
	vi.RegisterTagNameFunc(getLabelTagName)
	// 注册自定义验证规则 "httpmethod"
	vi.RegisterValidation("httpmethod", httpmethod)

	// 创建一个新的国际化实例
	eni := en.New()
	// 创建一个新的翻译器实例
	uti := ut.New(eni)
	// 获取英文翻译器
	ti, _ = uti.GetTranslator("en")

	// 注册默认的翻译
	err = ent.RegisterDefaultTranslations(vi, ti)
	checkErr(err)

	// 遍历默认标签数组
	for _, defaultTag := range defaultTags {
		// 注册翻译
		_ = registerTranslation(defaultTag, vi, ti, false)
	}
}

// Verify 根据validate标签验证结构体的可导出字段的数据合法性
func Verify(obj interface{}) (ves ValidateErrors) {
	defer func() {
		if c := recover(); c != nil {
			ves = append(ves, fmt.Sprintf("%s", c))
		}
	}()

	err := vi.Struct(obj)
	if err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			return ves
		}
		for _, err := range err.(validator.ValidationErrors) {
			ves = append(ves, err.Translate(ti))
		}
		return ves
	}

	return
}

// getLabelTagName 获取label标签名称
func getLabelTagName(sf reflect.StructField) string {
	name := strings.SplitN(sf.Tag.Get("label"), ",", 2)[0]
	if name == "-" {
		return ""
	} else if name == "" {
		return sf.Name
	}

	return name
}

// registerTranslation 注册翻译器
func registerTranslation(tag string, v *validator.Validate, t ut.Translator, override bool) error {
	return v.RegisterTranslation(tag, t, registerTranslationsFunc(tag, override), translationFunc(tag))
}

// registerTranslationsFunc 注册翻译装饰函数
func registerTranslationsFunc(key string, override bool) validator.RegisterTranslationsFunc {
	return func(ut ut.Translator) error {
		return ut.Add(key, "{0} field is invalid", override)
	}
}

// translationFunc 翻译装饰函数
func translationFunc(key string) validator.TranslationFunc {
	return func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T(key, fe.Field())
		return t
	}
}

// httpmethod http方法校验器
func httpmethod(fl validator.FieldLevel) bool {
	_, ok := httpMethodMap[strings.ToUpper(fl.Field().String())]
	return ok
}

// checkErr 检查错误
func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
