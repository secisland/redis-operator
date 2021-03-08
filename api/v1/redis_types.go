/*
Copyright 2021.

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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RedisSpec defines the desired state of Redis
type RedisSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Mode              string                     `json:"mode"`
	Size              *int32                     `json:"size,omitempty"`
	GlobalConfig      GlobalConfig               `json:"global"`
	Service           Service                    `json:"service"`
	Master            RedisMaster                `json:"master,omitempty"`
	Slave             RedisSlave                 `json:"slave,omitempty"`
	RedisExporter     *RedisExporter             `json:"redisExporter,omitempty"`
	RedisConfig       map[string]string          `json:"redisConfig"`
	Resources         *Resources                 `json:"resources,omitempty"`
	Storage           *Storage                   `json:"storage,omitempty"`
	NodeSelector      map[string]string          `json:"nodeSelector,omitempty"`
	SecurityContext   *corev1.PodSecurityContext `json:"securityContext,omitempty"`
	PriorityClassName string                     `json:"priorityClassName,omitempty"`
	Affinity          *corev1.Affinity           `json:"affinity,omitempty"`
}

type GlobalConfig struct {
	Image           string            `json:"image"`
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
	Password        *string           `json:"password,omitempty"`
	Resources       *Resources        `json:"resources,omitempty"`
}

type Resources struct {
	ResourceRequests ResourceDescription `json:"requests,omitempty"`
	ResourceLimits   ResourceDescription `json:"limits,omitempty"`
}

type ResourceDescription struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}

type Service struct {
	Type string `json:"type"`
}

type RedisMaster struct {
	Resources   Resources         `json:"resource,omitempty"`
	RedisConfig map[string]string `json:"redisConfig,omitempty"`
	Service     Service           `json:"service,omitempty"`
}

type RedisSlave struct {
	Resources   Resources         `json:"resource,omitempty"`
	RedisConfig map[string]string `json:"redisConfig,omitempty"`
	Service     Service           `json:"service,omitempty"`
}

type RedisExporter struct {
	Enabled         bool              `json:"enabled,omitempty"`
	Image           string            `json:"image"`
	Resources       *Resources        `json:"resources,omitempty"`
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
}

type Storage struct {
	VolumeClaimTemplate corev1.PersistentVolumeClaim `json:"volumeClaimTemplate,omitempty"`
}

// RedisStatus defines the observed state of Redis
type RedisStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Cluster RedisSpec `json:"cluster,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Redis is the Schema for the redis API
type Redis struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RedisSpec   `json:"spec,omitempty"`
	Status RedisStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RedisList contains a list of Redis
type RedisList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Redis `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Redis{}, &RedisList{})
}
