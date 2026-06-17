package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/ginutils"

	apiresponse "git.superwhys.top/rent-manager/rent-manager-server/api/response"
	apprentbill "git.superwhys.top/rent-manager/rent-manager-server/internal/app/rentbill"
	"git.superwhys.top/rent-manager/rent-manager-server/internal/xerr"
)

// RentBillRouter 注册租金账单路由。
func RentBillRouter(rentBillService *apprentbill.Service) ginutils.Option {
	return ginutils.WithGroupHandlers(
		ginutils.WithHandler(http.MethodGet, "/:rent_bill_id", GetRentBillHandler(rentBillService)),
		ginutils.WithHandler(http.MethodPost, "/:rent_bill_id/paid", MarkRentBillPaidHandler(rentBillService)),
		ginutils.WithHandler(http.MethodPost, "/:rent_bill_id/cancel", CancelRentBillHandler(rentBillService)),
	)
}

// ListRoomRentBillsHandler godoc
// @Summary 获取房间账单列表
// @Tags rent_bill
// @Produce json
// @Param room_id path int true "房间ID"
// @Param period_month query string false "账单月份 YYYY-MM"
// @Param status query string false "账单状态"
// @Param limit query int false "分页数量"
// @Param offset query int false "分页偏移"
// @Success 200 {object} ginutils.Ret[[]rentbill.RentBillResponse]
// @Security BearerAuth
// @Router /rooms/{room_id}/rent-bills [get]
func ListRoomRentBillsHandler(rentBillService *apprentbill.Service) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *apprentbill.RoomRentBillsQuery) {
		bills, err := rentBillService.ListRoomRentBills(c.Request.Context(), req)
		if apiresponse.HandleRouterError(c, err, "list room rent bills failed", xerr.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, apiresponse.ResponseWithData(apprentbill.ToRentBillListResponse(bills)))
	})
}

// GenerateRoomRentBillHandler godoc
// @Summary 生成房间月度账单
// @Tags rent_bill
// @Accept json
// @Produce json
// @Param room_id path int true "房间ID"
// @Param request body rentbill.GenerateRoomRentBillCommand true "生成账单参数"
// @Success 200 {object} ginutils.Ret[rentbill.GeneratedRentBillsResponse]
// @Security BearerAuth
// @Router /rooms/{room_id}/rent-bills/generate [post]
func GenerateRoomRentBillHandler(rentBillService *apprentbill.Service) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *apprentbill.GenerateRoomRentBillCommand) {
		result, err := rentBillService.GenerateRoomRentBill(c.Request.Context(), req)
		if apiresponse.HandleRouterError(c, err, "generate room rent bill failed", xerr.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, apiresponse.ResponseWithData(apprentbill.ToGeneratedRentBillsResponse(result)))
	})
}

// ListUnitRentBillsHandler godoc
// @Summary 获取单元账单列表
// @Tags rent_bill
// @Produce json
// @Param unit_id path int true "单元ID"
// @Param period_month query string false "账单月份 YYYY-MM"
// @Param status query string false "账单状态"
// @Param limit query int false "分页数量"
// @Param offset query int false "分页偏移"
// @Success 200 {object} ginutils.Ret[[]rentbill.RentBillResponse]
// @Security BearerAuth
// @Router /units/{unit_id}/rent-bills [get]
func ListUnitRentBillsHandler(rentBillService *apprentbill.Service) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *apprentbill.UnitRentBillsQuery) {
		bills, err := rentBillService.ListUnitRentBills(c.Request.Context(), req)
		if apiresponse.HandleRouterError(c, err, "list unit rent bills failed", xerr.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, apiresponse.ResponseWithData(apprentbill.ToRentBillListResponse(bills)))
	})
}

// GenerateUnitRentBillsHandler godoc
// @Summary 批量生成单元月度账单
// @Tags rent_bill
// @Accept json
// @Produce json
// @Param unit_id path int true "单元ID"
// @Param request body rentbill.GenerateUnitRentBillsCommand true "生成账单参数"
// @Success 200 {object} ginutils.Ret[rentbill.GeneratedRentBillsResponse]
// @Security BearerAuth
// @Router /units/{unit_id}/rent-bills/generate [post]
func GenerateUnitRentBillsHandler(rentBillService *apprentbill.Service) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *apprentbill.GenerateUnitRentBillsCommand) {
		result, err := rentBillService.GenerateUnitRentBills(c.Request.Context(), req)
		if apiresponse.HandleRouterError(c, err, "generate unit rent bills failed", xerr.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, apiresponse.ResponseWithData(apprentbill.ToGeneratedRentBillsResponse(result)))
	})
}

// GetRentBillHandler godoc
// @Summary 获取账单详情
// @Tags rent_bill
// @Produce json
// @Param rent_bill_id path int true "账单ID"
// @Success 200 {object} ginutils.Ret[rentbill.RentBillResponse]
// @Security BearerAuth
// @Router /rent-bills/{rent_bill_id} [get]
func GetRentBillHandler(rentBillService *apprentbill.Service) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *apprentbill.RentBillIDQuery) {
		bill, err := rentBillService.GetRentBill(c.Request.Context(), req.RentBillID)
		if apiresponse.HandleRouterError(c, err, "get rent bill failed", xerr.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, apiresponse.ResponseWithData(apprentbill.ToRentBillResponse(bill)))
	})
}

// MarkRentBillPaidHandler godoc
// @Summary 标记账单已收款
// @Tags rent_bill
// @Accept json
// @Produce json
// @Param rent_bill_id path int true "账单ID"
// @Param request body rentbill.MarkRentBillPaidCommand true "收款参数"
// @Success 200 {object} ginutils.Ret[rentbill.RentBillResponse]
// @Security BearerAuth
// @Router /rent-bills/{rent_bill_id}/paid [post]
func MarkRentBillPaidHandler(rentBillService *apprentbill.Service) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *apprentbill.MarkRentBillPaidCommand) {
		bill, err := rentBillService.MarkRentBillPaid(c.Request.Context(), req)
		if apiresponse.HandleRouterError(c, err, "mark rent bill paid failed", xerr.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, apiresponse.ResponseWithData(apprentbill.ToRentBillResponse(bill)))
	})
}

// CancelRentBillHandler godoc
// @Summary 取消账单
// @Tags rent_bill
// @Accept json
// @Produce json
// @Param rent_bill_id path int true "账单ID"
// @Param request body rentbill.CancelRentBillCommand true "取消参数"
// @Success 200 {object} ginutils.Ret[rentbill.RentBillResponse]
// @Security BearerAuth
// @Router /rent-bills/{rent_bill_id}/cancel [post]
func CancelRentBillHandler(rentBillService *apprentbill.Service) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *apprentbill.CancelRentBillCommand) {
		bill, err := rentBillService.CancelRentBill(c.Request.Context(), req)
		if apiresponse.HandleRouterError(c, err, "cancel rent bill failed", xerr.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, apiresponse.ResponseWithData(apprentbill.ToRentBillResponse(bill)))
	})
}
