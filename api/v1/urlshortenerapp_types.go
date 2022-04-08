/*
Copyright 2022.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// UrlShortenerAppSpec defines the desired state of UrlShortenerApp
type UrlShortenerAppSpec struct {
	//+kubebuilder:validation:Minimum=0
	Size int32 `json:"size"`

	// +kubebuilder:default=Always
	ImagePullPolicy string `json:"image-pull-policy,omitempty"`

	// +kubebuilder:default=80
	ContainerPort int32 `json:"container-port,omitempty"`
}

type UrlShortenerAppStatus struct {
	Nodes []string `json:"nodes"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// UrlShortenerApp is the Schema for the urlshortenerapps API
type UrlShortenerApp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UrlShortenerAppSpec   `json:"spec,omitempty"`
	Status UrlShortenerAppStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// UrlShortenerAppList contains a list of UrlShortenerApp
type UrlShortenerAppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []UrlShortenerApp `json:"items"`
}

func init() {
	SchemeBuilder.Register(&UrlShortenerApp{}, &UrlShortenerAppList{})
}
