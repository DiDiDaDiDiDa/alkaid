/*
 *  Copyright 2020. The Alkaid Authors. All rights reserved.
 *  Use of this source code is governed by a MIT-style
 *  license that can be found in the LICENSE file.
 *  Alkaid is a BaaS service based on Hyperledger Fabric.
 */

package organization

import (
	"github.com/gin-gonic/gin"
	"github.com/yakumioto/glog"
)

const (
	organizationID     = "organization_id"
	organizationDetail = "/:" + organizationID
)

var (
	logger *glog.Logger
)

type Handler struct{}

func (h *Handler) Router(e *gin.Engine) {
	logger = glog.MustGetLogger("handler.organization")
	defer logger.Infof("Organization handles initialization success.")

	r := e.Group("/organization")
	r.POST("/")
	r.GET("/")
	r.GET(organizationDetail)
	r.PATCH(organizationDetail)
	r.DELETE(organizationDetail)
}
