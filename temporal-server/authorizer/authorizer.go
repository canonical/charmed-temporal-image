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
	"strings"

	"go.temporal.io/server/common/authorization"
)

type authorizer struct{}

// NewAuthorizer returns a new authorization.Authorizer implementation.
func NewAuthorizer() authorization.Authorizer {
	return &authorizer{}
}

var decisionAllow = authorization.Result{Decision: authorization.DecisionAllow}
var decisionDeny = authorization.Result{Decision: authorization.DecisionDeny}

// Authorize returns an authorization decision (either DecisionAllow or
// DecisionDeny) based on the information contained in the provided Claims.
//
// It determines if the request is for a read or write operation and it checks
// the namespace to which the request is directed, comparing that against the
// claims to take a decision.
//
// Note: the provided Claims are trusted completely, no additional checks are
// performed on their source.
func (a *authorizer) Authorize(_ context.Context, claims *authorization.Claims,
	target *authorization.CallTarget) (authorization.Result, error) {

	if claims == nil {
		return decisionDeny, nil
	}

	requiredRole := authorization.RoleWriter
	if authorization.IsReadOnlyGlobalAPI(shortApiName(target.APIName)) {
		requiredRole = authorization.RoleReader
	} else if authorization.IsReadOnlyNamespaceAPI(shortApiName(target.APIName)) {
		requiredRole = authorization.RoleReader
	}

	if claims.System >= requiredRole {
		return decisionAllow, nil
	}

	if claims.Namespaces[target.Namespace] >= requiredRole {
		return decisionAllow, nil
	}

	return decisionDeny, nil
}

func shortApiName(api string) string {
	index := strings.LastIndex(api, "/")
	if index > -1 {
		return api[index+1:]
	}
	return api
}
