package logging

import (
	"context"
	"testing"
)

func TestNew(t *testing.T) {
	logger, err := New()

	if logger == nil || err != nil {
		t.Errorf("Logger: %#v \n Err: %#v", logger, err)
	}
}

func TestFromContext(t *testing.T) {
	want := &Logger{}
	ctx := NewContext(context.Background(), want)
	if got, ok := FromContext(ctx); got != want || !ok {
		t.Errorf("FromContext(%v) = %v, %v; want %v false", ctx, got, ok, want)
	}
}
