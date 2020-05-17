/*
 * Copyright 2020. The Alkaid Authors. All rights reserved.
 * Use of this source code is governed by a MIT-style
 * license that can be found in the LICENSE file.
 *
 * Alkaid is a BaaS service based on Hyperledger Fabric.
 *
 */

package types

const (
	DockerNetworkType      = "docker"
	DockerSwarmNetworkType = "docker_swarm"
	KubernetesNetworkType  = "kubernetes"
)

type Network struct {
	ID              int64  `json:"-"`
	NetworkID       string `json:"network_id,omitempty" binding:"required"`
	Name            string `json:"name,omitempty"`
	Type            string `json:"type,omitempty" binding:"required"`
	Description     string `json:"description,omitempty"`
	DockerNetworkID string `json:"-"`
	CreatedAt       int64  `json:"created_at,omitempty"`
	UpdatedAt       int64  `json:"updated_at,omitempty"`
}

func NewNetwork() *Network {
	return &Network{}
}

func (n *Network) GetNetworkID() string {
	switch n.Type {
	case DockerNetworkType:
		return n.DockerNetworkID
	default:
		return ""
	}
}
