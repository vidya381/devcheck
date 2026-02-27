package check

import (
	"context"
	"testing"
)

func TestBinaryCheck_Pass(t *testing.T) {
	c := &BinaryCheck{Binary: "go"}
	result := c.Run(context.Background())
	if result.Status != StatusPass {
		t.Errorf("expected pass, got %v: %s", result.Status, result.Message)
	}
}

func TestBinaryCheck_Fail(t *testing.T) {
	c := &BinaryCheck{Binary: "definitelynotabinary12345"}
	result := c.Run(context.Background())
	if result.Status != StatusFail {
		t.Errorf("expected fail, got %v", result.Status)
	}
}
