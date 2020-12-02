module github.com/khos2ow/provider-digitalocean

go 1.13

require (
	github.com/crossplane/crossplane-runtime v0.11.1-0.20201120062856-57ef784bfe43
	github.com/crossplane/crossplane-tools v0.0.0-20201007233256-88b291e145bb
	github.com/digitalocean/godo v1.54.0
	github.com/google/go-cmp v0.5.0
	github.com/pkg/errors v0.9.1
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	k8s.io/api v0.18.8
	k8s.io/apimachinery v0.18.8
	sigs.k8s.io/controller-runtime v0.6.2
	sigs.k8s.io/controller-tools v0.4.0
)
