/*
 * Copyright 2020. The Alkaid Authors. All rights reserved.
 * Use of this source code is governed by a MIT-style
 * license that can be found in the LICENSE file.
 *
 * Alkaid is a BaaS service based on Hyperledger Fabric.
 *
 */

package user

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yakumioto/glog"

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

var (
	logger *glog.Logger
)

type Service struct{}

func (s *Service) Init(log *glog.Logger, rg *gin.RouterGroup) {
	logger = log.MustGetLogger("user")
	r := rg.Group("/user")
	r.POST("", s.CreateUser)
	r.GET("")
	r.GET(userDetail, s.GetUserByID)
	r.PATCH(userDetail)
	r.DELETE(userDetail)

	logger.Infof("Service initialization success.")
}

func (s *Service) CreateUser(ctx *gin.Context) {
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

		ctx.Status(http.StatusInternalServerError)
		return
	}

	if err := s.EnrollCertificate(user, org); err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}

	if err := db.CreateMSP((*db.MSP)(user)); err != nil {
		var exist *db.ErrMSPExist
		if errors.As(err, &exist) {
			ctx.JSON(http.StatusBadRequest, apierrors.NewErrors(apierrors.DataAlreadyExists))
			return
		}

		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (s *Service) GetUserByID(ctx *gin.Context) {
	orgid := ctx.Param("organizationID")
	userid := ctx.Param("userID")

	msp, err := db.QueryMSPByOrganizationIDAndUserID(orgid, userid)
	if err != nil {
		var notExist *db.ErrMSPNotExist
		if errors.As(err, &notExist) {
			ctx.JSON(http.StatusNotFound, apierrors.NewErrors(apierrors.DataNotExists))
			return
		}

		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, msp)
}

func (s *Service) EnrollCertificate(user *types.User, org *types.Organization) error {
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
