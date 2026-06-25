package validators

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// slugPattern matches a public tenant slug: lowercase letters, digits and
// hyphens, 3–63 chars, not starting/ending with a hyphen.
var slugPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,61}[a-z0-9]$`)

func init() {
	_ = validate.RegisterValidation("slug", func(fl validator.FieldLevel) bool {
		return slugPattern.MatchString(fl.Field().String())
	})
}

func Struct(v any) error {
	return validate.Struct(v)
}
