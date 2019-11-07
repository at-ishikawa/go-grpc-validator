package grpc_playground_validator

import (
	"context"
	"reflect"
	"sort"
	"testing"

	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/fr"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	validatorv9 "gopkg.in/go-playground/validator.v9"
	fr_translations "gopkg.in/go-playground/validator.v9/translations/fr"
)

func TestNewValidator(t *testing.T) {
	got, gotErr := NewValidator()
	if got == nil {
		t.Errorf("want non nil, got nil")
	}
	if gotErr != nil {
		t.Error(gotErr)
	}
}

func TestValidator_ValidateGRPCRequest(t *testing.T) {
	type innerField struct {
		Field int `validate:"min=100"`
	}

	type request struct {
		RequiredField    string `validate:"required"`
		NonRequiredField string `validate:"max=8"`
		NonTaggedField   string
		InnerField       innerField
	}

	defaultValidator, err := NewValidator()
	if err != nil {
		t.Fatal(err)
	}
	translatedValidator, err := NewValidator(
		WithTranslators(en.New(), fr.New()),
		WithRegisterDefaultTranslationFunc(fr.New().Locale(), fr_translations.RegisterDefaultTranslations),
	)
	if err != nil {
		t.Fatal(err)
	}

	englishStatus, err := status.New(codes.InvalidArgument, "failed to validate request").
		WithDetails(&errdetails.BadRequest{
			FieldViolations: []*errdetails.BadRequest_FieldViolation{
				{
					Field:       "request.InnerField.Field",
					Description: "Field must be 100 or greater",
				},
				{
					Field:       "request.RequiredField",
					Description: "RequiredField is a required field",
				},
			}})
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		name      string
		validator *Validator
		request   interface{}
		locale    string

		want    *status.Status
		wantErr error
	}{
		{
			name:      "request is invalid",
			validator: defaultValidator,
			request:   &request{},
			locale:    "",
			want:      englishStatus,
			wantErr:   nil,
		},
		{
			name:      "request is invalid and different translator",
			validator: translatedValidator,
			request:   &request{},
			locale:    fr.New().Locale(),
			want: func() *status.Status {
				st, err := status.New(codes.InvalidArgument, "failed to validate request").
					WithDetails(&errdetails.BadRequest{
						FieldViolations: []*errdetails.BadRequest_FieldViolation{
							{
								Field:       "request.InnerField.Field",
								Description: "Field doit \303\252tre \303\251gal \303\240 100 ou plus",
							},
							{
								Field:       "request.RequiredField",
								Description: "RequiredField est un champ obligatoire",
							},
						}})
				if err != nil {
					t.Fatal(err)
				}
				return st
			}(),
			wantErr: nil,
		},
		{
			name:      "request is valid",
			validator: defaultValidator,
			request: &request{
				RequiredField: "required",
				InnerField: innerField{
					Field: 100,
				},
			},
			locale:  en.New().Locale(),
			want:    nil,
			wantErr: nil,
		},
		{
			name:      "locale is invalid",
			validator: defaultValidator,
			request:   &request{},
			locale:    "unknown locale",
			want:      englishStatus,
			wantErr:   nil,
		},
		{
			name:      "request is nil",
			validator: defaultValidator,
			locale:    "",
			request:   nil,
			want:      nil,
			wantErr:   nil,
		},
		{
			name:      "request parameter is not validatable",
			validator: defaultValidator,
			request:   "request",
			want:      nil,
			wantErr: &validatorv9.InvalidValidationError{
				Type: reflect.TypeOf("request"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := NewContextWithLocale(context.Background(), tc.locale)
			got, gotErr := tc.validator.ValidateGRPCRequest(ctx, tc.request)
			if got != nil && got.Details() != nil && len(got.Details()) > 0 {
				br, _ := got.Details()[0].(*errdetails.BadRequest)
				fvs := br.FieldViolations
				sort.Slice(fvs, func(i, j int) bool {
					return fvs[i].Field < fvs[j].Field
				})
				br.FieldViolations = fvs
				got, err = status.New(got.Code(), got.Message()).WithDetails(br)
				if err != nil {
					t.Fatal(err)
				}
			}
			if !reflect.DeepEqual(tc.want, got) {
				t.Errorf("want (message: %s, details: %+v), got (message: %s, details: %+v)", tc.want.Message(), tc.want.Details(), got.Message(), got.Details())
			}
			if !reflect.DeepEqual(tc.wantErr, gotErr) {
				t.Errorf("want %#v, got %#v", tc.wantErr, gotErr)
			}
		})
	}
}
