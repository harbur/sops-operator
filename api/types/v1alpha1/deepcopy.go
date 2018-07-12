package v1alpha1

import "k8s.io/apimachinery/pkg/runtime"

// DeepCopyInto copies all properties of this object into another object of the
// same type that is provided as a pointer.
func (in *SealedSecret) DeepCopyInto(out *SealedSecret) {
	out.TypeMeta = in.TypeMeta
	out.ObjectMeta = in.ObjectMeta
	out.Spec = SealedSecretSpec{
		Replicas: in.Spec.Replicas,
		Owner:    in.Spec.Owner,
	}
}

// DeepCopyObject returns a generically typed copy of an object
func (in *SealedSecret) DeepCopyObject() runtime.Object {
	out := SealedSecret{}
	in.DeepCopyInto(&out)

	return &out
}

// DeepCopyObject returns a generically typed copy of an object
func (in *SealedSecretList) DeepCopyObject() runtime.Object {
	out := SealedSecretList{}
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta

	if in.Items != nil {
		out.Items = make([]SealedSecret, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}

	return &out
}
