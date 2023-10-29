package authorizer_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/canonical/charmed-temporal-image/temporal-server/authorizer"
	mock "github.com/canonical/charmed-temporal-image/temporal-server/authorizer/mocks"
	"github.com/golang/mock/gomock"

	qt "github.com/frankban/quicktest"
	"go.temporal.io/server/common/authorization"
)

func TestTokenVerifier(t *testing.T) {
	c := qt.New(t)

	validToken := &authorizer.TokenInfo{
		Exp:           fmt.Sprint(time.Now().Add(time.Hour).Unix()),
		EmailVerified: "true",
		Email:         "user@example.com",
		Scope:         "https://www.googleapis.com/auth/userinfo.email",
		Azp:           "client_id",
	}

	expiredToken := &authorizer.TokenInfo{
		Exp:           fmt.Sprint(time.Now().Add(-time.Hour).Unix()),
		EmailVerified: "true",
		Email:         "user@example.com",
		Scope:         "https://www.googleapis.com/auth/userinfo.email",
		Azp:           "client_id",
	}

	emailNotVerifiedToken := &authorizer.TokenInfo{
		Exp:           fmt.Sprint(time.Now().Add(time.Hour).Unix()),
		EmailVerified: "false",
		Email:         "user@example.com",
		Scope:         "https://www.googleapis.com/auth/userinfo.email",
		Azp:           "client_id",
	}

	noScopeToken := &authorizer.TokenInfo{
		Exp:           fmt.Sprint(time.Now().Add(time.Hour).Unix()),
		EmailVerified: "true",
		Email:         "user@example.com",
		Scope:         "profile",
		Azp:           "client_id",
	}

	differentClientToken := &authorizer.TokenInfo{
		Exp:           fmt.Sprint(time.Now().Add(time.Hour).Unix()),
		EmailVerified: "true",
		Email:         "user@example.com",
		Scope:         "https://www.googleapis.com/auth/userinfo.email",
		Azp:           "badwolf_client_id",
	}

	serviceAccountToken := &authorizer.TokenInfo{
		Exp:           fmt.Sprint(time.Now().Add(time.Hour).Unix()),
		EmailVerified: "true",
		Email:         "service-account@project-id.iam.gserviceaccount.com",
		Scope:         "https://www.googleapis.com/auth/userinfo.email",
		Azp:           "123",
	}

	tests := []struct {
		desc string
		// Inputs
		token *authorizer.TokenInfo
		// Outputs
		expectedErr string
	}{
		{
			desc:  "success: valid token",
			token: validToken,
		},
		{
			desc:        "error: expired token",
			token:       expiredToken,
			expectedErr: "token expired",
		},
		{
			desc:        "error: token email not verified",
			token:       emailNotVerifiedToken,
			expectedErr: "token email not verified",
		},
		{
			desc:        "error: missing required scope",
			token:       noScopeToken,
			expectedErr: "token scope must include email",
		},
		{
			desc:        "error: mismatch in client id",
			token:       differentClientToken,
			expectedErr: "incorrect token client id",
		},
		{
			desc:  "success: valid service account token",
			token: serviceAccountToken,
		},
	}

	for _, test := range tests {
		test := test

		c.Run(test.desc, func(c *qt.C) {
			c.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			tv := authorizer.NewVerifier("client_id", "https://www.googleapis.com/oauth2/v3/tokeninfo", "https://www.googleapis.com/auth/userinfo.email")

			err := tv.VerifyToken(test.token)
			if test.expectedErr != "" {
				c.Assert(err, qt.ErrorMatches, test.expectedErr)
			} else {
				c.Assert(err, qt.IsNil)
			}
		})
	}
}

func TestGetClaims(t *testing.T) {
	c := qt.New(t)

	validToken := &authorizer.TokenInfo{
		Exp:           fmt.Sprint(time.Now().Add(time.Hour).Unix()),
		EmailVerified: "true",
		Email:         "user@example.com",
		Scope:         "https://www.googleapis.com/auth/userinfo.email",
	}
	validAuthToken := "Bearer W3siYyI6W3siaSI6InRpbWUtYmVmb3JlIDIwMjItMTEtMDJUMDA6MDA6MDBaIn0seyJpIjoidXNlcm5hbWUgbG9seW91Zm91bmRtZSJ9XSwibCI6ImFnZW50IiwiaTY0IjoiQXdvUVFGbDA5VmFFVTBQT1RlRlRhQU9Rd3hJZ056bGxOMkpsTjJJeE9EVTJNRGczWldKbFpEZ3haVE0xWW1VNU1qZGtNek1hRGdvRmJHOW5hVzRTQld4dloybHUiLCJzNjQiOiJDeWJ2bTY3Sjc3N2UxeVBTS0U0RWE5dDgtQzdsRnJiZ3RVTmhudUVfc3R3In1dCg=="
	tests := []struct {
		desc string
		// Inputs
		authInfo          *authorization.AuthInfo
		adminGroups       string
		setupExpectations func(tv *mock.MockTokenVerifier, np *mock.MockNamespaceAccessProvider) []*gomock.Call
		// Outputs
		expectedClaims *authorization.Claims
		expectedErr    string
	}{{
		desc: "success: authInfo contains valid token and user belongs to admin group",
		authInfo: &authorization.AuthInfo{
			AuthToken: validAuthToken,
		},
		adminGroups: "system",
		setupExpectations: func(tv *mock.MockTokenVerifier, np *mock.MockNamespaceAccessProvider) []*gomock.Call {
			return []*gomock.Call{
				tv.EXPECT().GetTokenInfo(gomock.Any()).Return(validToken, nil),
				tv.EXPECT().VerifyToken(gomock.Any()).Return(nil),
				np.EXPECT().GetNamespaceAccessInformation(gomock.Any(), gomock.Any()).Return([]authorizer.NamespaceAccess{{Namespace: "foobar", Relation: "writer"}}, nil),
			}
		},
		expectedClaims: &authorization.Claims{
			Namespaces: map[string]authorization.Role{"": authorization.RoleReader, "foobar": authorization.RoleWriter},
		},
	},
		{
			desc: "success: authInfo contains valid token and user belongs to admin group and specific group",
			authInfo: &authorization.AuthInfo{
				AuthToken: validAuthToken,
			},
			adminGroups: "system",
			setupExpectations: func(tv *mock.MockTokenVerifier, np *mock.MockNamespaceAccessProvider) []*gomock.Call {
				return []*gomock.Call{
					tv.EXPECT().GetTokenInfo(gomock.Any()).Return(validToken, nil),
					tv.EXPECT().VerifyToken(gomock.Any()).Return(nil),
					np.EXPECT().GetNamespaceAccessInformation(gomock.Any(), gomock.Any()).Return([]authorizer.NamespaceAccess{{Namespace: "foobar", Relation: "writer"}, {Namespace: "system", Relation: "admin"}}, nil),
				}
			},
			expectedClaims: &authorization.Claims{
				System:     authorization.RoleWriter,
				Namespaces: map[string]authorization.Role{"": authorization.RoleReader, "foobar": authorization.RoleWriter},
			},
		},
	}

	for _, test := range tests {
		test := test

		c.Run(test.desc, func(c *qt.C) {
			c.Parallel()

			// Execution
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			tv := mock.NewMockTokenVerifier(ctrl)
			np := mock.NewMockNamespaceAccessProvider(ctrl)
			if test.setupExpectations != nil {
				gomock.InAnyOrder(test.setupExpectations(tv, np))
			}

			cm := authorizer.TokenClaimMapper{
				TokenVerifier:           tv,
				NamespaceAccessProvider: np,
				AdminGroups:             test.adminGroups,
			}
			claims, err := cm.GetClaims(test.authInfo)
			c.Assert(claims, qt.DeepEquals, test.expectedClaims)
			if test.expectedErr != "" {
				c.Assert(err, qt.ErrorMatches, test.expectedErr)
			}
		})
	}
}
