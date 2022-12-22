package authenticator_test

import (
	"context"
	td "github.com/fatmalabidi/MailBoxServcie/src/testdata/authenticator"
	"github.com/fatmalabidi/MailBoxServcie/src/utils/helpers"
	"testing"
)

func Test_LoadCredentials(t *testing.T) {
	t.Parallel()
	for _, testcase := range td.CreateTTLoadCredentials() {
		t.Run(testcase.Name, func(t *testing.T) {
			ctx, cancel := helpers.AddTimeoutToCtx(context.Background(), 10)
			defer cancel()
			err := testcase.Auth.LoadCredentials(ctx)
			if err == nil && testcase.HasError {
				t.Errorf("expected error, got nothing")
			}
			if !testcase.HasError {
				if err != nil {
					t.Errorf("expected success, got error %v", err)
				}
				userName, password := helpers.GetCredFromEnv()
				if userName != td.Tusername || password != td.Tpasword {
					t.Errorf("expected %v got %v", td.Tusername, userName)
				}
			}
		})
	}
}
