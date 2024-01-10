// The MIT License
//
// Copyright (c) 2020 Temporal Technologies Inc.  All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package authorizer

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/canonical/ofga"
	"go.temporal.io/server/common/authorization"
	"go.uber.org/zap"
)

//go:generate mockgen -destination=mocks/groups_provider_gen.go -package=mock github.com/canonical/charmed-temporal-image/temporal-server/authorizer NamespaceAccessProvider,TokenVerifier

// TokenVerifier is an interface that defines the methods
// to fetch token information and verify their validity.
type TokenVerifier interface {
	GetTokenInfo(accessToken string) (*TokenInfo, error)
	VerifyToken(token *TokenInfo) error
}

type NamespaceAccess struct {
	Namespace string
	Relation  string
}

// NamespaceAccessProvider is an interface that defines the method to retrieve namespace
// access information for a given email.
type NamespaceAccessProvider interface {
	GetUserGroups(ctx context.Context, email string) ([]string, error)
	GetNamespaceAccessInformation(ctx context.Context, email string, groups []string) ([]NamespaceAccess, error)
}

// AuthClient implements the necessary methods needed to fetch namespace access
// information from an OpenFGA store.
type AuthClient struct {
	OfgaClient *ofga.Client
}

// TokenClaimMapper implements Temporal authorization.ClaimMapper,
// performing a OIDC-based authorization.
type TokenClaimMapper struct {
	// TokenVerifier is used to verify the provided access token.
	TokenVerifier TokenVerifier
	// NamespaceAccessProvider is used to identify the namespaces that the user
	// logging in via the access token has access to.
	NamespaceAccessProvider NamespaceAccessProvider
	// AdminGroups is a comma-separated list of groups which gives full system
	// access to all users belonging to them.
	AdminGroups string
	// OpenAccessNamespaces is a comma-separated list of namespaces to which everyone
	// with valid login credentials has access. If empty, no such namespace will be
	// configured.
	OpenAccessNamespaces string
	// Logger is used for logging TokenClaimMapper operations.
	Logger *zap.Logger
}

var roleMap = map[string]authorization.Role{
	"reader": authorization.RoleReader,
	"writer": authorization.RoleWriter,
	"admin":  authorization.RoleAdmin,
}

func NewTokenClaimMapper(ctx context.Context, cfg *ConfigWithAuth, logger *zap.Logger) (authorization.ClaimMapper, error) {
	client, err := ofga.NewClient(ctx, ofga.OpenFGAParams{
		Scheme:      cfg.Auth.OFGA.APIScheme,
		Host:        cfg.Auth.OFGA.APIHost,
		Port:        cfg.Auth.OFGA.APIPort,
		Token:       cfg.Auth.OFGA.BearerToken,
		StoreID:     cfg.Auth.OFGA.StoreID,
		AuthModelID: cfg.Auth.OFGA.AuthModelID,
	})
	if err != nil {
		return nil, fmt.Errorf("error connecting to ofga client: %v", err)
	}
	return &TokenClaimMapper{
		NamespaceAccessProvider: &AuthClient{OfgaClient: client},
		TokenVerifier:           NewVerifier(cfg.Auth.GoogleClientID, "https://www.googleapis.com/oauth2/v3/tokeninfo", "https://www.googleapis.com/auth/userinfo.email"),
		Logger:                  logger,
		AdminGroups:             cfg.Auth.AdminGroups,
		OpenAccessNamespaces:    cfg.Auth.OpenAccessNamespaces,
	}, nil
}

// GetClaims implements authorization.ClaimMapper.GetClaims. It expects the
// AuthInfo.AuthToken (received from the `Authorization` header of the request)
// to be in the format of `Bearer <token>` where `<token>` is a valid
// Google IAM access token.
//
// It then verifies the groups that the user presented in the access token belongs
// to via OpenFGA and gives access to various Temporal namespaces according to them.
//
// If the user belongs to any of the AdminGroups groups, they get RoleWriter on the global
// System namespace. If the user is a member of a group in OpenFGA with some level of
// access to a given namespace, they get that access level to the namespace. E.g.
// If user `john` is a member of group `abc`, and group `abc` is related to namespace `example`
// as a "writer", then user `john` will be assigned RoleWriter on namespace `example`.
// Additionally, they get RoleReader on empty namespace in order to perform initiating
// calls required by the SDK.
func (c TokenClaimMapper) GetClaims(authInfo *authorization.AuthInfo) (*authorization.Claims, error) {
	claims := authorization.Claims{
		Namespaces: make(map[string]authorization.Role),
	}

	if authInfo.AuthToken == "" {
		return nil, errors.New("no auth token provided")
	}

	token := strings.TrimPrefix(authInfo.AuthToken, "Bearer ")
	if len(token) == len(authInfo.AuthToken) {
		return nil, errors.New("invalid token length")
	}

	tokenInfo, err := c.TokenVerifier.GetTokenInfo(token)
	if err != nil {
		return nil, c.generateError(fmt.Sprintf("error fetching access token info: %v", err))
	}

	err = c.TokenVerifier.VerifyToken(tokenInfo)
	if err != nil {
		return nil, c.generateError(fmt.Sprintf("error validating access token: %v", err))
	}

	email := tokenInfo.Email
	ctx := context.Background()

	adminGroupsSlice := strings.Split(c.AdminGroups, ",")
	userGroups, err := c.NamespaceAccessProvider.GetUserGroups(ctx, email)
	if err != nil {
		return nil, c.generateError(fmt.Sprintf("error reading group membership: %v \n", err))
	}

	// Check for admin group membership
	for _, grp := range adminGroupsSlice {
		if grp != "" && slices.Contains(userGroups, grp) {
			claims.System = authorization.RoleWriter
			return &claims, nil
		}
	}

	namespaceAccess, err := c.NamespaceAccessProvider.GetNamespaceAccessInformation(ctx, email, userGroups)
	if err != nil {
		return nil, c.generateError(fmt.Sprintf("error reading namespace access: %v \n", err))
	}

	openAccessNamespaces := strings.Split(c.OpenAccessNamespaces, ",")
	for _, ns := range openAccessNamespaces {
		claims.Namespaces[ns] = authorization.RoleWriter
	}

	hasNamespaces := false
	for _, ns := range namespaceAccess {
		if ns.Namespace != "" {
			role, exists := roleMap[ns.Relation]
			if exists {
				claims.Namespaces[ns.Namespace] = role
				hasNamespaces = true
			}
		}
	}

	if len(claims.Namespaces) > 0 {
		claims.Namespaces[""] = authorization.RoleReader
	}

	if !hasNamespaces && c.Logger != nil {
		c.Logger.Warn(fmt.Sprintf("received request with valid token but no namespace access; groups found: %v", userGroups))
	}

	return &claims, nil
}

// GetNamespaceAccessInformation returns a list of namespaces that a user with the given email
// has access to along with the type of relation they have (One of "reader", "writer" or "admin").
func (c *AuthClient) GetNamespaceAccessInformation(ctx context.Context, email string, groups []string) ([]NamespaceAccess, error) {
	var namespaceAccess []NamespaceAccess
	for _, group := range groups {
		continuationToken := ""

		for {
			tuples, continuationToken, err := c.OfgaClient.FindMatchingTuples(ctx, ofga.Tuple{
				Object:   &ofga.Entity{Kind: "group", ID: fmt.Sprintf("%v#member", group)},
				Relation: "",
				Target:   &ofga.Entity{Kind: "namespace", ID: ""},
			}, 100, continuationToken)
			if err != nil {
				return nil, err
			}

			for _, tuple := range tuples {
				namespaceAccess = append(namespaceAccess, NamespaceAccess{
					Namespace: tuple.Tuple.Target.ID,
					Relation:  tuple.Tuple.Relation.String(),
				})
			}

			if continuationToken == "" {
				break
			}
		}
	}

	return namespaceAccess, nil
}

// GetUserGroups returns a list of groups that a user with the given email
// is a member of in the OpenFGA store.
func (c *AuthClient) GetUserGroups(ctx context.Context, email string) ([]string, error) {
	var groups []string
	continuationToken := ""
	for {
		tuples, continuationToken, err := c.OfgaClient.FindMatchingTuples(ctx, ofga.Tuple{
			Object:   &ofga.Entity{Kind: "user", ID: email},
			Relation: "",
			Target:   &ofga.Entity{Kind: "group", ID: ""},
		}, 100, continuationToken)
		if err != nil {
			return nil, err
		}

		for _, tuple := range tuples {
			groups = append(groups, tuple.Tuple.Target.ID)
		}

		if continuationToken == "" {
			break
		}
	}

	return groups, nil
}

// generateError returns a new error and also logs it on the provided logger.
func (c TokenClaimMapper) generateError(msg string) error {
	if c.Logger != nil {
		c.Logger.Error(msg)
	}
	return errors.New(msg)
}
