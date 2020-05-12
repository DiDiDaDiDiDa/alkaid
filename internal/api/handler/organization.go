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
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	apierrors "github.com/yakumioto/alkaid/internal/api/errors"
	"github.com/yakumioto/alkaid/internal/api/types"
	"github.com/yakumioto/alkaid/internal/db"
	"github.com/yakumioto/alkaid/internal/utils/certificate"
)

const (
	organizationID     = "orgnaization_id"
	organizationDetail = "/:" + organizationID
)

type Organization struct{}

func (o *Organization) Init(e *gin.Engine) {
	r := e.Group("/organization")
	r.POST("", o.CreateOrganization)
	r.GET("")
	r.GET(organizationDetail, o.GetOrganizationByID)
	r.PATCH(organizationDetail)
	r.DELETE(organizationDetail)

	logger.Infof("Organization handles initialization success.")
}

func (o *Organization) CreateOrganization(ctx *gin.Context) {
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
		returnInternalServerError(ctx, "New sign CA error: %v", err)
		return
	}
	org.TLSCAPrivateKey, org.TLSCACertificate, err = certificate.NewCA(pkixName)
	if err != nil {
		returnInternalServerError(ctx, "New TLS CA error: %v", err)
		return
	}

	if err := db.CreateOrganization((*db.Organization)(org)); err != nil {
		var exist *db.ErrOrganizationExist
		if errors.As(err, &exist) {
			ctx.JSON(http.StatusBadRequest, apierrors.NewErrors(apierrors.DataAlreadyExists))
			return
		}

		returnInternalServerError(ctx, "Insert organization error: %s", err)
		return
	}

	ctx.JSON(http.StatusOK, org)
}

func (o *Organization) GetOrganizationByID(ctx *gin.Context) {
	id := ctx.Param("organizationID")

	org, err := db.QueryOrganizationByOrgID(id)
	if err != nil {
		var notExist *db.ErrOrganizationNotExist
		if errors.As(err, &notExist) {
			ctx.JSON(http.StatusBadRequest, apierrors.NewErrors(apierrors.DataNotExists))
			return
		}

		returnInternalServerError(ctx, "Query organization by organization_id error: %v", err)
		return
	}

	ctx.JSON(http.StatusOK, org)
}
