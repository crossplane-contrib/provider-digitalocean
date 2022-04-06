module github.com/crossplane-contrib/provider-digitalocean

go 1.16

require (
	github.com/alecthomas/units v0.0.0-20210927113745-59d0afb8317a // indirect
	github.com/aws/aws-sdk-go-v2 v1.16.1
	github.com/aws/aws-sdk-go-v2/config v1.15.2
	github.com/aws/aws-sdk-go-v2/service/s3 v1.26.0
	github.com/crossplane/crossplane-runtime v0.15.1
	github.com/crossplane/crossplane-tools v0.0.0-20210916125540-071de511ae8e
	github.com/digitalocean/godo v1.77.0
	github.com/golang/mock v1.5.0
	github.com/google/go-cmp v0.5.7
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/pkg/errors v0.9.1
	golang.org/x/net v0.0.0-20220127200216-cd36cc0744dd // indirect
	golang.org/x/oauth2 v0.0.0-20211005180243-6b3c2da341f1 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	k8s.io/api v0.22.2
	k8s.io/apimachinery v0.22.2
	sigs.k8s.io/controller-runtime v0.9.2
	sigs.k8s.io/controller-tools v0.7.0
)

replace github.com/googleapis/gnostic v0.5.6 => github.com/google/gnostic v0.5.6
