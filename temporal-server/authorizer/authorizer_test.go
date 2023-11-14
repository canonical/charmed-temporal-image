package authorizer_test

import (
	"context"
	"testing"

	"github.com/canonical/charmed-temporal-image/temporal-server/authorizer"

	qt "github.com/frankban/quicktest"
	"go.temporal.io/server/common/authorization"
)

func TestAuthorize(t *testing.T) {
	c := qt.New(t)

	tests := []struct {
		desc string
		// Inputs
		claims   *authorization.Claims
		target   string
		targetNS string
		// Outputs
		expectedDecision authorization.Decision
	}{{
		desc: "allow: caller has system read access permissions, read-only API",
		claims: &authorization.Claims{
			System: authorization.RoleReader,
		},
		target:           "ListNamespaces",
		expectedDecision: authorization.DecisionAllow,
	}, {
		desc: "deny: caller has system read access permissions, write API",
		claims: &authorization.Claims{
			System: authorization.RoleReader,
		},
		target:           "CreateNamespace",
		expectedDecision: authorization.DecisionDeny,
	}, {
		desc: "allow: caller has system write access permissions",
		claims: &authorization.Claims{
			System: authorization.RoleWriter,
		},
		target:           "CreateNamespace",
		expectedDecision: authorization.DecisionAllow,
	}, {
		desc: "allow: caller has specific namespace permissions, get system info",
		claims: &authorization.Claims{
			Namespaces: map[string]authorization.Role{"": authorization.RoleReader, "test-ns": authorization.RoleWriter},
		},
		target:           "GetSystemInfo",
		expectedDecision: authorization.DecisionAllow,
	}, {
		desc: "deny: caller has specific namespace permissions, list namespaces",
		claims: &authorization.Claims{
			Namespaces: map[string]authorization.Role{"": authorization.RoleReader, "test-ns": authorization.RoleWriter},
		},
		target:           "CreateNamespace",
		expectedDecision: authorization.DecisionDeny,
	}, {
		desc: "allow: caller has specific namespace permissions, execute workflow",
		claims: &authorization.Claims{
			Namespaces: map[string]authorization.Role{"": authorization.RoleReader, "test-ns": authorization.RoleWriter},
		},
		target:           "StartWorkflow",
		targetNS:         "test-ns",
		expectedDecision: authorization.DecisionAllow,
	}, {
		desc: "deny: caller has specific namespace permissions, execute workflow in another ns",
		claims: &authorization.Claims{
			Namespaces: map[string]authorization.Role{"": authorization.RoleReader, "test-ns": authorization.RoleWriter},
		},
		target:           "StartWorkflow",
		targetNS:         "test-ns-2",
		expectedDecision: authorization.DecisionDeny,
	}, {
		desc: "allow: caller has system write access permissions, execute workflow",
		claims: &authorization.Claims{
			System: authorization.RoleWriter,
		},
		target:           "StartWorkflow",
		targetNS:         "test-ns",
		expectedDecision: authorization.DecisionAllow,
	}}

	for _, test := range tests {
		c.Run(test.desc, func(c *qt.C) {
			c.Parallel()

			// Execution
			a := authorizer.NewAuthorizer()
			result, err := a.Authorize(context.Background(), test.claims, &authorization.CallTarget{
				APIName:   test.target,
				Namespace: test.targetNS,
			})

			c.Assert(result.Decision, qt.Equals, test.expectedDecision)
			c.Assert(err, qt.IsNil)
		})
	}
}
