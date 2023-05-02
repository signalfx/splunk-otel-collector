package utils

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func FindOwnerWithUID(ors []metav1.OwnerReference, uid types.UID) *metav1.OwnerReference {
	for _, or := range ors {
		if or.UID == uid {
			return &or
		}
	}
	return nil
}

func FindOwnerWithKind(ors []metav1.OwnerReference, kind string) *metav1.OwnerReference {
	for _, or := range ors {
		if or.Kind == kind {
			return &or
		}
	}
	return nil
}
