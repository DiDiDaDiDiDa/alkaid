/*
 * Copyright 2020. The Alkaid Authors. All rights reserved.
 * Use of this source code is governed by a MIT-style
 * license that can be found in the LICENSE file.
 * Alkaid is a BaaS service based on Hyperledger Fabric.
 */

package cryptoconfig

import (
	"fmt"

	"github.com/yakumioto/alkaid/internal/api/types"
	"github.com/yakumioto/alkaid/internal/utils/certificate"
	"github.com/yakumioto/alkaid/internal/utils/targz"
	"github.com/yakumioto/alkaid/third_party/github.com/hyperledger/fabric/common/crypto"
)

func GetMSPArchive(org *types.Organization, user *types.User) ([]byte, error) {
	priv, err := certificate.Signer(user.SignPrivateKey)
	if err != nil {
		return nil, err
	}

	cacertFileName := fmt.Sprintf("cacerts/ca.%s-cert.pem", org.Domain)

	archive := targz.New()
	archive.
		AddFile("config.yaml",
			0664,
			[]byte(fmt.Sprintf(`NodeOUs:
  Enable: true
  ClientOUIdentifier:
    Certificate: %s
    OrganizationalUnitIdentifier: client
  PeerOUIdentifier:
    Certificate: %s
    OrganizationalUnitIdentifier: peer
  AdminOUIdentifier:
    Certificate: %s
    OrganizationalUnitIdentifier: admin
  OrdererOUIdentifier:
    Certificate: %s
    OrganizationalUnitIdentifier: orderer`, cacertFileName, cacertFileName, cacertFileName, cacertFileName))).
		AddFile(
			cacertFileName,
			0664,
			org.CACertificate).
		AddFile(
			fmt.Sprintf("tlscacerts/tlsca.%s-cert.pem", org.Domain),
			0664,
			org.TLSCACertificate).
		AddFile(
			fmt.Sprintf("keystore/%x_sk", crypto.ComputeSKI(priv.PrivateKey)),
			0600,
			user.SignPrivateKey).
		AddFile(
			fmt.Sprintf("signcerts/%s.%s-cert.pem", user.UserID, org.Domain),
			0664,
			user.SignCertificate)

	data, err := archive.Generate()
	if err != nil {
		return nil, err
	}
	return data, nil
}

func GetTLSArchive(org *types.Organization, user *types.User) ([]byte, error) {
	archive := targz.New()
	archive.
		AddFile(
			"ca.crt",
			0664,
			org.TLSCACertificate).
		AddFile(
			"server.crt",
			0664,
			user.TLSCertificate).
		AddFile(
			"server.key",
			0600,
			user.TLSPrivateKey)

	data, err := archive.Generate()
	if err != nil {
		return nil, err
	}
	return data, nil
}
