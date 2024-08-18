/*
Copyright 2024.

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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// EnvironmentSpec defines the desired state of Environment
type EnvironmentSpec struct {
	ComputeService ComputeService `json:"computeService"`
	StorageService StorageService `json:"storageService"`
}

// CPUUnit represents the number of CPU cores
type CPUUnit int

// MemoryUnit represents a memory size with units MiB or GiB
type MemoryUnit struct {
	Value int    `json:"value"`
	Unit  string `json:"unit"` // "MiB" or "GiB"
}

// Server represents a server with CPU and memory specifications
type Server struct {
	ServerType string     `json:"serverType"`
	CPU        CPUUnit    `json:"cpu"`
	Memory     MemoryUnit `json:"memory"`
}

// ComputeService represents a service that manages compute resources
type ComputeService struct {
	NumServers int      `json:"numServers"`
	Servers    []Server `json:"servers"`
}

// StorageUnit represents a storage size with units MiB or GiB
type StorageUnit struct {
	Value int    `json:"value"`
	Unit  string `json:"unit"` // "MiB" or "GiB"
}

// StorageType represents different types of storage (e.g., SSD, HDD)
// +kubebuilder:validation:Enum=SSD;HDD
type StorageType string
type ServerType string

const (
	SSD      StorageType = "SSD"
	HDD      StorageType = "HDD"
	T2Micros ServerType  = "t2.micro"
)

// StorageService represents a service that manages storage resources
type StorageService struct {
	StorageType StorageType `json:"storageType"`
	Capacity    StorageUnit `json:"capacity"`
}

/*
Current Number of Servers: Reflect the actual number of servers currently provisioned and running.
Available Memory and CPU: Show the total available memory and CPU resources across all servers.
Storage Utilization: Indicate how much of the storage capacity is currently in use.
Conditions: Include conditions to represent the state of the environment (e.g., "Ready", "Degraded", "Error").
Last Update Time: Track the last time the status was updated.
*/

// EnvironmentStatus defines the observed state of Environment
type EnvironmentStatus struct {
	CurrentNumServers  int                `json:"currentNumServers"`
	AvailableCPU       CPUUnit            `json:"availableCPU"`
	AvailableMemory    MemoryUnit         `json:"availableMemory"`
	StorageUtilization StorageUnit        `json:"storageUtilization"`
	Conditions         []metav1.Condition `json:"conditions"`
}

// // Condition represents the state of the resource at a certain point.
// type Condition struct {
// 	Type               string      `json:"type"`
// 	Status             string      `json:"status"`
// 	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
// 	Reason             string      `json:"reason"`
// 	Message            string      `json:"message"`
// }

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Environment is the Schema for the environments API
type Environment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EnvironmentSpec   `json:"spec,omitempty"`
	Status EnvironmentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// EnvironmentList contains a list of Environment
type EnvironmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Environment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Environment{}, &EnvironmentList{})
}
