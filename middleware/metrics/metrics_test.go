package metrics

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMetrics(t *testing.T) {
	next := func(ctx context.Context, req interface{}) (interface{}, error) {
		return req.(string) + "https://go-kratos.dev", nil
	}
	_, err := Server()(next)(context.Background(), "test:")
	assert.Equal(t, err, nil)

	_, err = Client()(next)(context.Background(), "test:")
	assert.Equal(t, err, nil)
}
