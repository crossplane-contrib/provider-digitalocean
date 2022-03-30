package storage

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/crossplane-contrib/provider-digitalocean/apis/storage/v1alpha1"
)

// GenerateSpace creates a s3.CreateBucketInput from a given DOSpaceParameter
func GenerateSpace(name string, in v1alpha1.DOSpaceParameters, create *s3.CreateBucketInput) {
	create.Bucket = aws.String(name)
	create.ACL = s3types.BucketCannedACL(aws.ToString(in.ACL))
	create.GrantFullControl = in.GrantFullControl
	create.GrantRead = in.GrantRead
	create.GrantReadACP = in.GrantReadACP
	create.GrantWrite = in.GrantWrite
	create.GrantWriteACP = in.GrantWriteACP

	if in.ObjectOwnership != nil {
		create.ObjectOwnership = *in.ObjectOwnership
	}
}
