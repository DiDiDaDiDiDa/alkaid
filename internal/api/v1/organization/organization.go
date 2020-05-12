/*
 * Copyright 2020. The Alkaid Authors. All rights reserved.
 * Use of this source code is governed by a MIT-style
 * license that can be found in the LICENSE file.
 *
 * Alkaid is a BaaS service based on Hyperledger Fabric.
 *
 */

package organization

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/yakumioto/glog"

	apierrors "github.com/yakumioto/alkaid/internal/api/errors"
	"github.com/yakumioto/alkaid/internal/api/types"
	"github.com/yakumioto/alkaid/internal/db"
	"github.com/yakumioto/alkaid/internal/utils/certificate"
)

const (
	organizationID     = "orgnaization_id"
	organizationDetail = "/:" + organizationID
)

var (
	logger *glog.Logger
)

type Service struct{}

func (s *Service) Init(log *glog.Logger, rg *gin.RouterGroup) {
	logger = log.MustGetLogger("organization")

	r := rg.Group("/organization")
	r.POST("", s.CreateOrganization)
	r.GET("")
	r.GET(organizationDetail, s.GetOrganizationByID)
	r.PATCH(organizationDetail)
	r.DELETE(organizationDetail)

	logger.Infof("Service initialization success.")
}

func (s *Service) CreateOrganization(ctx *gin.Context) {
	org := types.NewOrganization()

	err := ctx.ShouldBindJSON(org)
	if err != nil {
		logger.Debuf("Bind JSON error: %v", err)
		ctx.JSON(http.StatusBadRequest, apierrors.NewErrors(apierrors.BadRequestData))
		return
	}

	pkixName := &certificate.PkixName{
		OrgName:       org.OrganizationID,
		Country:       org.Country,
		Province:      org.Province,
		Locality:      org.Locality,
		OrgUnit:       org.OrganizationalUnit,
		StreetAddress: org.StreetAddress,
		PostalCode:    org.PostalCode,
	}
	org.SignCAPrivateKey, org.SignCACertificate, err = certificate.NewCA(pkixName)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}
	org.TLSCAPrivateKey, org.TLSCACertificate, err = certificate.NewCA(pkixName)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}

	if err := db.CreateOrganization((*db.Organization)(org)); err != nil {
		var exist *db.ErrOrganizationExist
		if errors.As(err, &exist) {
			ctx.JSON(http.StatusBadRequest, apierrors.NewErrors(apierrors.DataAlreadyExists))
			return
		}

		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, org)
}

func (s *Service) GetOrganizationByID(ctx *gin.Context) {
	id := ctx.Param("organizationID")

	org, err := db.QueryOrganizationByOrgID(id)
	if err != nil {
		var notExist *db.ErrOrganizationNotExist
		if errors.As(err, &notExist) {
			ctx.JSON(http.StatusBadRequest, apierrors.NewErrors(apierrors.DataNotExists))
			return
		}

		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, org)
}
