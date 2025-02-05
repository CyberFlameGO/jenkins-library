package validation

import (
	"errors"
	"reflect"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	valid "github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

type Translation struct {
	Tag           string
	RegisterFn    valid.RegisterTranslationsFunc
	TranslationFn valid.TranslationFunc
}

type validation struct {
	Validator  *valid.Validate
	Translator ut.Translator
}

type validationOption func(*validation) error

func New(opts ...validationOption) (*validation, error) {
	validator := valid.New()
	enTranslator := en.New()
	universalTranslator := ut.New(enTranslator, enTranslator)
	translator, found := universalTranslator.GetTranslator("en")
	if !found {
		return nil, errors.New("translator for en locale is not found")
	}

	validation := &validation{
		Validator:  validator,
		Translator: translator,
	}

	for _, opt := range opts {
		if err := opt(validation); err != nil {
			return nil, err
		}
	}

	return validation, nil
}

func WithJSONNamesForStructFields() validationOption {
	return func(v *validation) error {
		v.Validator.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			return name
		})
		return nil
	}
}

func WithPredefinedErrorMessages() validationOption {
	translations := []Translation{
		{
			Tag: "oneof",
			RegisterFn: func(ut ut.Translator) error {
				return ut.Add("oneof", "The {0} must use the following values: {1}. ", true)
			},
			TranslationFn: func(ut ut.Translator, fe valid.FieldError) string {
				t, _ := ut.T("oneof", fe.Field(), fe.Param())
				return t
			},
		}, {
			Tag: "required_if",
			RegisterFn: func(ut ut.Translator) error {
				// TODO: Improve the message for condition required_if for several fields
				return ut.Add("required_if", "The {0} is required since the {1} is {2}. ", true)
			},
			TranslationFn: func(ut ut.Translator, fe valid.FieldError) string {
				params := []string{fe.Field()}
				params = append(params, strings.Split(fe.Param(), " ")...)
				t, _ := ut.T("required_if", params...)
				return t
			},
		},
	}
	return func(v *validation) error {
		if err := registerTranslations(translations, v.Validator, v.Translator); err != nil {
			return err
		}
		return nil
	}
}

func WithCustomErrorMessages(translations []Translation) validationOption {
	return func(v *validation) error {
		if err := registerTranslations(translations, v.Validator, v.Translator); err != nil {
			return err
		}
		return nil
	}
}

func (v *validation) ValidateStruct(s interface{}) error {
	var errStr string
	errs := v.Validator.Struct(s)
	if errs != nil {
		if err, ok := errs.(*valid.InvalidValidationError); ok {
			return err
		}
		for _, err := range errs.(valid.ValidationErrors) {
			errStr += err.Translate(v.Translator)
		}
		return errors.New(errStr)
	}
	return nil
}

func registerTranslations(translations []Translation, validator *valid.Validate, translator ut.Translator) error {
	if err := en_translations.RegisterDefaultTranslations(validator, translator); err != nil {
		return err
	}

	for _, t := range translations {
		if err := validator.RegisterTranslation(t.Tag, translator, t.RegisterFn, t.TranslationFn); err != nil {
			return err
		}
	}
	return nil
}
