package internal

import (
    `reflect`
    `testing`

    `google.golang.org/genproto/googleapis/rpc/errdetails`
)

func TestConvertValidationErrors(t *testing.T) {
    testCases := []struct{
        name string
        errorMessages map[string]string
        want *errdetails.BadRequest
    } {
        {
            name: "no error messages",
            errorMessages: nil,
            want: nil,
        },
        {
            name: "error messages in the same fields",
            errorMessages: map[string]string{
                "field1": "description1",
                "field2": "description2",
            },
            want: &errdetails.BadRequest{
                FieldViolations: []*errdetails.BadRequest_FieldViolation{
                    {
                        Field:                "field1",
                        Description:          "description1",
                    },
                    {
                        Field:                "field2",
                        Description:          "description2",
                    },
                },
            },
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            got := ConvertValidationErrors(tc.errorMessages)
            if !reflect.DeepEqual(tc.want, got) {
                t.Errorf("want %v, got %v", tc.want, got)
            }
        })
    }
}
