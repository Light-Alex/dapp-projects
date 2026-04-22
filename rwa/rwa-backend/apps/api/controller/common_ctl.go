package controller


import (
	"github.com/AnchoredLabs/rwa-backend/libs/core/web"
	"github.com/gin-gonic/gin"
)

type CommonController struct {
}

func NewCommonController() *CommonController {
	return &CommonController{}
}

// HealthCheck godoc
// @Summary HealthCheck
// @Description HealthCheck
// @Id HealthCheck
// @Tags Common
// @Produce  json
// @Success 200 {object} string "success"
// @Failure	401	{object} web.Response "Unauthorized"
// @Failure	500	{object} web.Response "Internal Server Error"
// @Router /common/health [get]
func (ctl *CommonController) HealthCheck(g *gin.Context) {
	web.ResponseOk("ok", g)
}
