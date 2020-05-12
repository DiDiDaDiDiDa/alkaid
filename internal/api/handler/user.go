/*
 * Copyright 2020. The Alkaid Authors. All rights reserved.
 * Use of this source code is governed by a MIT-style
 * license that can be found in the LICENSE file.
 *
 * Alkaid is a BaaS service based on Hyperledger Fabric.
 *
 */

package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	apierrors "github.com/yakumioto/alkaid/internal/api/errors"
	"github.com/yakumioto/alkaid/internal/api/types"
	"github.com/yakumioto/alkaid/internal/db"
	"github.com/yakumioto/alkaid/internal/utils/certificate"
	"github.com/yakumioto/alkaid/third_party/github.com/hyperledger/fabric/common/crypto"
)

const (
	userID     = "user_id"
	userDetail = "/:" + userID
)

type User struct{}

func (u *User) Init(e *gin.Engine) {
	r := e.Group("/user")
	r.POST("", u.CreateUser)
	r.GET("")
	r.GET(userDetail, u.GetUserByID)
	r.PATCH(userDetail)
	r.DELETE(userDetail)

	logger.Infof("User handles initialization success.")
}

func (u *User) CreateUser(ctx *gin.Context) {
	user := types.NewUser()
	if err := ctx.ShouldBindJSON(user); err != nil {
		logger.Debuf("Bind JSON error: %v", err)

		ctx.JSON(http.StatusBadRequest, apierrors.NewErrors(apierrors.BadRequestData))
		return
	}

	org, err := db.QueryOrganizationByOrgID(user.OrganizationID)
	if err != nil {
		var notExist *db.ErrOrganizationNotExist
		if errors.As(err, &notExist) {
			ctx.JSON(http.StatusNotFound, apierrors.NewErrors(apierrors.DataNotExists))
			return
		}

		returnInternalServerError(ctx, "Get Organization error: %v", err)
		return
	}

	if err := u.EnrollCertificate(user, org); err != nil {
		returnInternalServerError(ctx, "User initialize error: %s", err)
		return
	}

	if err := db.CreateMSP((*db.MSP)(user)); err != nil {
		var exist *db.ErrMSPExist
		if errors.As(err, &exist) {
			ctx.JSON(http.StatusBadRequest, apierrors.NewErrors(apierrors.DataAlreadyExists))
			return
		}

		returnInternalServerError(ctx, "Insert User error: %s", err)
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (u *User) GetUserByID(ctx *gin.Context) {
	orgid := ctx.Param("organizationID")
	userid := ctx.Param("userID")

	msp, err := db.QueryMSPByOrganizationIDAndUserID(orgid, userid)
	if err != nil {
		var notExist *db.ErrMSPNotExist
		if errors.As(err, &notExist) {
			ctx.JSON(http.StatusNotFound, apierrors.NewErrors(apierrors.DataNotExists))
			return
		}

		returnInternalServerError(ctx, "Query User by organization_id and user_id error: %v", err)
		return
	}

	ctx.JSON(http.StatusOK, msp)
}

func (u *User) EnrollCertificate(user *types.User, org *types.Organization) error {
	priv, err := crypto.GeneratePrivateKey()
	if err != nil {
		return fmt.Errorf("generate private key error: %v", err)
	}

	pkixName := org.GetCertificatePkixName()

	signCert, err := certificate.SignSignCertificate(pkixName, []string{user.Type}, nil,
		&priv.PublicKey, org.SignCAPrivateKey, org.SignCACertificate)
	if err != nil {
		return fmt.Errorf("sign signature certificate error: %v", err)
	}
	tlsCert, err := certificate.SignTLSCertificate(pkixName, []string{user.Type}, user.SANS,
		&priv.PublicKey, org.SignCAPrivateKey, org.SignCACertificate)
	if err != nil {
		return fmt.Errorf("sign tls certificate error: %v", err)
	}

	privBytes, err := crypto.PrivateKeyExport(priv)
	if err != nil {
		return fmt.Errorf("private key export error: %v", err)
	}

	user.PrivateKey = privBytes
	user.SignCertificate = crypto.X509Export(signCert)
	user.TLSCertificate = crypto.X509Export(tlsCert)

	return nil
}
