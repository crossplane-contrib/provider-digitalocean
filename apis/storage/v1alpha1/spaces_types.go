/*
Copyright 2020 The Crossplane Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

type ObjectOwnershipType string

const (
	BucketOwnerPreferred ObjectOwnershipType = "BucketOwnerPreferred"
	ObjectWriter         ObjectOwnershipType = "ObjectWriter"
	BucketOwnerEnforced  ObjectOwnershipType = "BucketOwnerEnforced"
)

// DOSpaceParameters define the desired state of a DigitalOcean Space
// Most fields map directly to the AWS S3 Go SDK
// See docs https://docs.aws.amazon.com/sdk-for-go/api/service/s3/#CreateBucketInput
type DOSpaceParameters struct {
	// The canned ACL to apply to the bucket.
	ACL *string `json:"acl,omitempty"`

	// The name of the bucket to create. This is required.
	BucketName string `json:"bucketName"`

	// Allows grantee the read, write, read ACP, and write ACP permissions on the bucket.
	// +kubebuilder:validation:Optional
	GrantFullControl *string `json:"grantFullControl,omitempty"`

	// Allows grantee to list the objects in the bucket.
	// +kubebuilder:validation:Optional
	GrantRead *string `json:"grantRead,omitempty"`

	// Allows grantee to read the bucket ACL.
	// +kubebuilder:validation:Optional
	GrantReadACP *string `json:"grantReadACP,omitempty"`

	// Allows grantee to create new objects in the bucket.
	// For the bucket and object owners of existing objects, also allows deletions and overwrites of those objects.
	// +kubebuilder:validation:Optional
	GrantWrite *string `json:"grantWrite,omitempty"`

	// Allows grantee to write the ACL for the applicable bucket.
	// +kubebuilder:validation:Optional
	GrantWriteACP *string `json:"grantWriteACP,omitempty"`

	// Specifies whether you want S3 Object Lock to be enabled for the new bucket.
	ObjectLockEnabledForBucket *bool `json:"objectLockEnabledForBucket,omitempty"`

	// The container element for object ownership for a bucket's ownership controls.
	//
	// BucketOwnerPreferred - Objects uploaded to the bucket change ownership to
	// the bucket owner if the objects are uploaded with the bucket-owner-full-control
	// canned ACL.
	//
	// ObjectWriter - The uploading account will own the object if the object is
	// uploaded with the bucket-owner-full-control canned ACL.
	//
	// BucketOwnerEnforced - Access control lists (ACLs) are disabled and no longer
	// affect permissions. The bucket owner automatically owns and has full control
	// over every object in the bucket. The bucket only accepts PUT requests that
	// don't specify an ACL or bucket owner full control ACLs, such as the bucket-owner-full-control
	// canned ACL or an equivalent form of this ACL expressed in the XML format.
	ObjectOwnership *ObjectOwnershipType `json:"objectOwnership,omitempty"`
}

type DOSpaceObservation struct{}

type DOSpaceSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       DOSpaceParameters `json:"forProvider"`
}

type DOSpaceStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          DOSpaceObservation `json:"atProvider,omitempty"`
}

type DOSpace struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DOSpaceSpec   `json:"spec"`
	Status DOSpaceStatus `json:"status,omitempty"`
}

type DOSpaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DOSpace `json:"items"`
}
