module github.com/go-kratos/kratos/config/kube/v2

go 1.15

require (
	github.com/go-kratos/kratos/v2 v2.0.0-rc3
	k8s.io/api v0.21.1
	k8s.io/apimachinery v0.21.1
	k8s.io/client-go v0.21.1
)

replace github.com/go-kratos/kratos/v2 => ../../../kratos
