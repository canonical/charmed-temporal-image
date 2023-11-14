package authorizer

import (
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestShortApiName(t *testing.T) {
	c := qt.New(t)

	tests := []struct {
		input    string
		expected string
	}{
		{input: "/temporal.api.workflowservice.v1.WorkflowService/StartWorkflowExecution", expected: "StartWorkflowExecution"},
		{input: "/StartWorkflowExecution", expected: "StartWorkflowExecution"},
		{input: "StartWorkflowExecution", expected: "StartWorkflowExecution"},
		{input: "", expected: ""},
	}

	for _, test := range tests {
		result := shortApiName(test.input)
		c.Assert(result, qt.Equals, test.expected)
	}
}
