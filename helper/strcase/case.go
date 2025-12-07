package strcase

import (
	"github.com/stoewer/go-strcase"
	"strings"
	"unicode"
)

func ToCamel(s string) string {
	return strcase.LowerCamelCase(s)
}

func ToKebab(s string) string {
	return strcase.KebabCase(s)
}

func ToPascal(s string) string {
	return strcase.UpperCamelCase(s)
}

func ToSnake(s string) string {
	return strcase.SnakeCase(s)
}

func ToScreamerSnake(s string) string {
	return strcase.UpperSnakeCase(s)
}

func ToSentence(s string) string {
	if len(s) < 1 {
		return ""
	}
	lowerCase := strings.ReplaceAll(strcase.SnakeCase(s), "_", " ")
	firstSymbol := []rune(lowerCase)[0]

	return string(unicode.ToUpper(firstSymbol)) + lowerCase[1:]
}
