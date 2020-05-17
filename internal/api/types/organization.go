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
	SignCAType = "sign"
	TLSCAType  = "tls"

	OrdererOrgType = "orderer"
	PeerOrgType    = "peer"
)

// Organization in the network
type Organization struct {
	ID             int64    `json:"-"`
	OrganizationID string   `json:"organization_id,omitempty" binding:"required"`
	Name           string   `json:"name,omitempty" binding:"required"`
	NetworkID      []string `json:"network_id,omitempty"`
	Domain         string   `json:"domain,omitempty" binding:"required,fqdn"`
	Description    string   `json:"description,omitempty"`

	// Type value is orderer or peer
	Type string `json:"type,omitempty" binding:"required,oneof=orderer peer"`

	// The following fields are the fields that generate the certificate
	Country            string `json:"country,omitempty"`
	Province           string `json:"province,omitempty"`
	Locality           string `json:"locality,omitempty"`
	OrganizationalUnit string `json:"organizational_unit,omitempty"`
	StreetAddress      string `json:"street_address,omitempty"`
	PostalCode         string `json:"postal_code,omitempty"`

	// sign and tls root ca
	SignCAPrivateKey []byte `json:"-"`
	TLSCAPrivateKey  []byte `json:"-"`
	CACertificate    []byte `json:"ca_certificate,omitempty"`
	TLSCACertificate []byte `json:"tlsca_certificate,omitempty"`

	CreateAt int64 `json:"create_at,omitempty"`
	UpdateAt int64 `json:"update_at,omitempty"`
}

// NewOrganization Default parameter
func NewOrganization() *Organization {
	return &Organization{
		Country:    "China",
		Province:   "Beijing",
		Locality:   "Beijing",
		PostalCode: "100000",
	}
}

func (o *Organization) HasNetwork(id string) bool {
	for _, nid := range o.NetworkID {
		if nid == id {
			return true
		}
	}
	return false
}
