package ecsagent

// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

// https://github.com/aws/amazon-ecs-agent/blob/6f0415557e583614dcd54c6547ef6ad01ade1a90/ecs-agent/tmds/handlers/response/response.go#L15

// VolumeResponse is the schema for the volume response JSON object
type VolumeResponse struct {
	DockerName  string `json:"DockerName,omitempty"`
	Source      string `json:"Source,omitempty"`
	Destination string `json:"Destination,omitempty"`
}

// PortResponse defines the schema for portmapping response JSON
// object.
type PortResponse struct {
	ContainerPort uint16 `json:"ContainerPort,omitempty"`
	Protocol      string `json:"Protocol,omitempty"`
	HostPort      uint16 `json:"HostPort,omitempty"`
	HostIp        string `json:"HostIp,omitempty"`
}

// Network is a struct that keeps track of metadata of a network interface
type Network struct {
	NetworkMode   string   `json:"NetworkMode,omitempty"`
	IPv4Addresses []string `json:"IPv4Addresses,omitempty"`
	IPv6Addresses []string `json:"IPv6Addresses,omitempty"`
}

// https://github.com/aws/amazon-ecs-agent/blob/6f0415557e583614dcd54c6547ef6ad01ade1a90/ecs-agent/stats/types.go
type NetworkStatsPerSec struct {
	RxBytesPerSecond float64 `json:"rx_bytes_per_sec"`
	TxBytesPerSecond float64 `json:"tx_bytes_per_sec"`
}
