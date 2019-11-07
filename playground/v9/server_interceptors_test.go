package grpc_playground_validator

import (
	"context"
	"reflect"
	"testing"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	validatorv9 "gopkg.in/go-playground/validator.v9"
	translationsen "gopkg.in/go-playground/validator.v9/translations/en"
)

func TestUnaryServerInterceptor(t *testing.T) {
	type request struct {
		RequiredField    string `validate:"required"`
		NonRequiredField string `validate:"max=8"`
		NonTaggedField   string
	}

	defaultValidator := &Validator{
		validate:            validatorv9.New(),
		universalTranslator: ut.New(en.New(), en.New()),
	}

	englishLocale := en.New()
	translatedValidator := &Validator{
		validate:            validatorv9.New(),
		universalTranslator: ut.New(englishLocale, englishLocale),
	}
	englishTranslator, ok := translatedValidator.universalTranslator.GetTranslator(englishLocale.Locale())
	if !ok {
		t.Fatal("want to get english translator")
	}
	err := translationsen.RegisterDefaultTranslations(translatedValidator.validate, englishTranslator)
	if err != nil {
		t.Fatal(err)
	}

	resp := "dummy"

	testCases := []struct {
		name      string
		validator *Validator

		request     interface{}
		response    interface{}
		responseErr *status.Status

		want    interface{}
		wantErr *status.Status
	}{
		{
			name:      "request is invalid",
			validator: defaultValidator,

			request:     &request{},
			response:    resp,
			responseErr: nil,
			want:        nil,
			wantErr: func() *status.Status {
				st, err := status.New(codes.InvalidArgument, "failed to validate request").
					WithDetails(&errdetails.BadRequest{
						FieldViolations: []*errdetails.BadRequest_FieldViolation{
							{
								Field:       "request.RequiredField",
								Description: "Key: 'request.RequiredField' Error:Field validation for 'RequiredField' failed on the 'required' tag",
							},
						}})
				if err != nil {
					t.Fatal(err)
				}
				return st
			}(),
		},
		{
			name:        "request is invalid and different translator",
			validator:   translatedValidator,
			request:     &request{},
			response:    resp,
			responseErr: nil,
			want:        nil,
			wantErr: func() *status.Status {
				st, err := status.New(codes.InvalidArgument, "failed to validate request").
					WithDetails(&errdetails.BadRequest{
						FieldViolations: []*errdetails.BadRequest_FieldViolation{
							{
								Field:       "request.RequiredField",
								Description: "RequiredField is a required field",
							},
						}})
				if err != nil {
					t.Fatal(err)
				}
				return st
			}(),
		},
		{
			name:      "request is valid",
			validator: defaultValidator,
			request: &request{
				RequiredField: "required",
			},
			response:    resp,
			responseErr: status.New(codes.Internal, "internal error"),
			want:        resp,
			wantErr:     status.New(codes.Internal, "internal error"),
		},
		{
			name:        "request is nil",
			validator:   defaultValidator,
			request:     nil,
			response:    resp,
			responseErr: nil,
			want:        resp,
			wantErr:     nil,
		},
		{
			name:        "request parameter is not validatable",
			validator:   defaultValidator,
			request:     "request",
			response:    resp,
			responseErr: nil,
			want:        nil,
			wantErr:     status.New(codes.FailedPrecondition, "request is not able to validate: validator: (nil string)"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			interceptor := UnaryServerInterceptor(tc.validator)
			got, gotErr := interceptor(context.Background(), tc.request, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
				return tc.response, tc.responseErr.Err()
			})
			if !reflect.DeepEqual(tc.want, got) {
				t.Errorf("want %#v, got %#v", tc.want, got)
			}
			if !reflect.DeepEqual(tc.wantErr.Err(), gotErr) {
				gotSt, ok := status.FromError(gotErr)
				if !ok {
					t.Error("got error should be *status.Status")
				}
				t.Errorf("want (message: %#v, details: %+v), got (message: %#v, details: %+v)", tc.wantErr.Message(), tc.wantErr.Details(), gotSt.Message(), gotSt.Details())
			}
		})
	}
}
