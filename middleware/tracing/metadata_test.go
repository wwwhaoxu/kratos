package tracing

import (
	"context"
	"testing"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/metadata"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/propagation"
)

func TestMetadata_Inject(t *testing.T) {
	type args struct {
		appName string
		carrier propagation.TextMapCarrier
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "https://go-kratos.dev",
			args: args{"https://go-kratos.dev", propagation.HeaderCarrier{}},
			want: "https://go-kratos.dev",
		},
		{
			name: "https://github.com/go-kratos/kratos",
			args: args{"https://github.com/go-kratos/kratos", propagation.HeaderCarrier{"mode": []string{"test"}}},
			want: "https://github.com/go-kratos/kratos",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := kratos.New(kratos.Name(tt.args.appName))
			ctx := kratos.NewContext(context.Background(), a)
			var m = new(Metadata)
			m.Inject(ctx, tt.args.carrier)
			if res := tt.args.carrier.Get(serviceHeader); tt.want != res {
				t.Errorf("Get(serviceHeader) :%s want: %s", res, tt.want)
			}
		})
	}
}

func TestMetadata_Extract(t *testing.T) {
	type args struct {
		parent  context.Context
		carrier propagation.TextMapCarrier
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "https://go-kratos.dev",
			args: args{
				parent:  context.Background(),
				carrier: propagation.HeaderCarrier{"X-Md-Service-Name": []string{"https://go-kratos.dev"}},
			},
			want: "https://go-kratos.dev",
		},
		{
			name: "https://github.com/go-kratos/kratos",
			args: args{
				parent:  metadata.NewServerContext(context.Background(), metadata.Metadata{}),
				carrier: propagation.HeaderCarrier{"X-Md-Service-Name": []string{"https://github.com/go-kratos/kratos"}},
			},
			want: "https://github.com/go-kratos/kratos",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := Metadata{}
			ctx := b.Extract(tt.args.parent, tt.args.carrier)
			md, ok := metadata.FromServerContext(ctx)
			assert.Equal(t, ok, true)
			assert.Equal(t, md.Get(serviceHeader), tt.want)
		})
	}
}

func TestFields(t *testing.T) {
	b := Metadata{}
	assert.Equal(t, b.Fields(), []string{"x-md-service-name"})
}
