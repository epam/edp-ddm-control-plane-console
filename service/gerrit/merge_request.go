package gerrit

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type GerritMergeRequest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              GerritMergeRequestSpec   `json:"spec,omitempty"`
	Status            GerritMergeRequestStatus `json:"status,omitempty"`
}

type GerritMergeRequestSpec struct {
	OwnerName           string   `json:"ownerName"`
	ProjectName         string   `json:"projectName"`
	TargetBranch        string   `json:"targetBranch"`
	SourceBranch        string   `json:"sourceBranch"`
	CommitMessage       string   `json:"commitMessage"`
	AuthorName          string   `json:"authorName"`
	AuthorEmail         string   `json:"authorEmail"`
	ChangesConfigMap    string   `json:"changesConfigMap"`
	AdditionalArguments []string `json:"additionalArguments"`
}

type GerritMergeRequestStatus struct {
	Value     string `json:"value"`
	ChangeURL string `json:"changeUrl"`
	ChangeID  string `json:"changeId"`
}

func (in GerritMergeRequest) FormattedCreatedAt() string {
	return in.CreationTimestamp.Format(ViewTimeFormat)
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type GerritMergeRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GerritMergeRequest `json:"items"`
}

func (in GerritMergeRequest) OwnerName() string {
	return in.Spec.OwnerName
}

func (in GerritMergeRequest) TargetBranch() string {
	if in.Spec.TargetBranch == "" {
		return "master"
	}

	return in.Spec.TargetBranch
}

func (in GerritMergeRequest) CommitMessage() string {
	if in.Spec.CommitMessage == "" && in.Spec.SourceBranch != "" {
		return fmt.Sprintf("merge %s to %s", in.Spec.SourceBranch, in.TargetBranch())
	} else if in.Spec.CommitMessage == "" && in.Spec.ChangesConfigMap != "" {
		return fmt.Sprintf("merge files contents from config map: %s", in.Spec.ChangesConfigMap)
	}

	return in.Spec.CommitMessage
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GerritMergeRequest) DeepCopyInto(out *GerritMergeRequest) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GerritMergeRequest.
func (in *GerritMergeRequest) DeepCopy() *GerritMergeRequest {
	if in == nil {
		return nil
	}
	out := new(GerritMergeRequest)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *GerritMergeRequest) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GerritMergeRequestList) DeepCopyInto(out *GerritMergeRequestList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]GerritMergeRequest, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GerritMergeRequestList.
func (in *GerritMergeRequestList) DeepCopy() *GerritMergeRequestList {
	if in == nil {
		return nil
	}
	out := new(GerritMergeRequestList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *GerritMergeRequestList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GerritMergeRequestSpec) DeepCopyInto(out *GerritMergeRequestSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GerritMergeRequestSpec.
func (in *GerritMergeRequestSpec) DeepCopy() *GerritMergeRequestSpec {
	if in == nil {
		return nil
	}
	out := new(GerritMergeRequestSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GerritMergeRequestStatus) DeepCopyInto(out *GerritMergeRequestStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GerritMergeRequestStatus.
func (in *GerritMergeRequestStatus) DeepCopy() *GerritMergeRequestStatus {
	if in == nil {
		return nil
	}
	out := new(GerritMergeRequestStatus)
	in.DeepCopyInto(out)
	return out
}
