package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/ginutils"
	"github.com/miebyte/goutils/logging"
	domainaccount "github.com/superwhys/rent-manager-server/internal/domain/account"
	domainmeteritem "github.com/superwhys/rent-manager-server/internal/domain/meteritem"
	domainmeterreading "github.com/superwhys/rent-manager-server/internal/domain/meterreading"
	domainrbac "github.com/superwhys/rent-manager-server/internal/domain/rbac"
	domainrentbill "github.com/superwhys/rent-manager-server/internal/domain/rentbill"
	domainroom "github.com/superwhys/rent-manager-server/internal/domain/room"
	domainroomtype "github.com/superwhys/rent-manager-server/internal/domain/roomtype"
	domainunit "github.com/superwhys/rent-manager-server/internal/domain/unit"
	"github.com/superwhys/rent-manager-server/internal/errcode"
)

// common.HandleRouterError 处理业务逻辑错误响应。
func HandleRouterError(ctx *gin.Context, err error, logMsg string, fallback errcode.Error) bool {
	if err == nil {
		return false
	}

	if errcode.IsAuthExpired(err) {
		ctx.JSON(http.StatusUnauthorized, ginutils.ErrorRet(err))
		return true
	}

	if ec, ok := errcode.AsErrcode(err); ok {
		ctx.JSON(http.StatusOK, ginutils.ErrorRet(ec))
		return true
	}
	if ec, ok := classifyDomainError(err); ok {
		ctx.JSON(http.StatusOK, ginutils.ErrorRet(ec.WithMessage(err.Error())))
		return true
	}

	logging.Errorw(ctx.Request.Context(), logMsg, logging.Err(err))
	ctx.JSON(http.StatusOK, ginutils.ErrorRet(fallback))
	return true
}

func classifyDomainError(err error) (errcode.Error, bool) {
	switch {
	case errors.Is(err, domainaccount.ErrAccountNotFound),
		errors.Is(err, domainunit.ErrNotFound),
		errors.Is(err, domainroom.ErrNotFound),
		errors.Is(err, domainroomtype.ErrNotFound),
		errors.Is(err, domainmeteritem.ErrNotFound),
		errors.Is(err, domainmeterreading.ErrNotFound),
		errors.Is(err, domainmeterreading.ErrRoomNotFound),
		errors.Is(err, domainmeterreading.ErrMeterItemNotFound),
		errors.Is(err, domainrentbill.ErrNotFound),
		errors.Is(err, domainrentbill.ErrRoomNotFound),
		errors.Is(err, domainrbac.ErrPermissionGroupNotFound),
		errors.Is(err, domainrbac.ErrPermissionNotFound),
		errors.Is(err, domainrbac.ErrRoleNotFound),
		errors.Is(err, domainrbac.ErrUnitGrantNotFound):
		return errcode.ErrNotFound, true
	case errors.Is(err, domainrbac.ErrForbidden),
		errors.Is(err, domainrbac.ErrUnitNotInAccount),
		errors.Is(err, domainrbac.ErrRoleNotInAccount):
		return errcode.ErrForbidden, true
	case isKnownDomainError(err):
		return errcode.ErrBadRequest, true
	default:
		return errcode.Error{}, false
	}
}

func isKnownDomainError(err error) bool {
	knownErrors := []error{
		domainaccount.ErrWechatOpenIDRequired,
		domainunit.ErrAccountRequired,
		domainunit.ErrNameRequired,
		domainunit.ErrAddressRequired,
		domainunit.ErrCityRequired,
		domainunit.ErrProvinceRequired,
		domainunit.ErrNameAlreadyExists,
		domainunit.ErrInvalidListScope,
		domainunit.ErrInvalidMeterItemID,
		domainunit.ErrMeterItemNotFound,
		domainunit.ErrUnitHasRooms,
		domainroom.ErrRoomNoRequired,
		domainroom.ErrInvalidFloor,
		domainroom.ErrInvalidPrice,
		domainroom.ErrInvalidSize,
		domainroom.ErrUnitRequired,
		domainroom.ErrRoomTypeRequired,
		domainroom.ErrInvalidMeterItemID,
		domainroom.ErrMeterItemNotFound,
		domainroomtype.ErrAccountRequired,
		domainroomtype.ErrNameRequired,
		domainroomtype.ErrNameTooLong,
		domainroomtype.ErrNameAlreadyExists,
		domainroomtype.ErrRoomTypeHasRooms,
		domainmeteritem.ErrAccountRequired,
		domainmeteritem.ErrCodeRequired,
		domainmeteritem.ErrUnitRequired,
		domainmeteritem.ErrInvalidUnitPrice,
		domainmeteritem.ErrInvalidCalcType,
		domainmeteritem.ErrCodeAlreadyExists,
		domainmeterreading.ErrRoomRequired,
		domainmeterreading.ErrMeterItemRequired,
		domainmeterreading.ErrPeriodMonthRequired,
		domainmeterreading.ErrInvalidPeriodMonth,
		domainmeterreading.ErrInvalidPreviousReading,
		domainmeterreading.ErrInvalidCurrentReading,
		domainmeterreading.ErrInvalidUsageAmount,
		domainmeterreading.ErrPreviousReadingRequired,
		domainmeterreading.ErrCurrentReadingDecreased,
		domainmeterreading.ErrReadingTimeRequired,
		domainmeterreading.ErrInvalidSource,
		domainmeterreading.ErrInvalidCalcType,
		domainmeterreading.ErrInvalidUnitPrice,
		domainmeterreading.ErrAlreadyExists,
		domainmeterreading.ErrRoomMeterItemNotAssigned,
		domainrentbill.ErrRoomRequired,
		domainrentbill.ErrPeriodMonthRequired,
		domainrentbill.ErrInvalidPeriodMonth,
		domainrentbill.ErrInvalidBaseRent,
		domainrentbill.ErrInvalidMeterFee,
		domainrentbill.ErrInvalidTotalAmount,
		domainrentbill.ErrInvalidStatus,
		domainrentbill.ErrPaidAtRequired,
		domainrentbill.ErrAlreadyPaid,
		domainrentbill.ErrCancelled,
		domainrentbill.ErrAlreadyExists,
		domainrbac.ErrRoleAccountRequired,
		domainrbac.ErrRoleNameRequired,
		domainrbac.ErrPermissionGroupCodeRequired,
		domainrbac.ErrPermissionGroupNameRequired,
		domainrbac.ErrPermissionCodeRequired,
		domainrbac.ErrPermissionDescriptionRequired,
		domainrbac.ErrPermissionGroupAlreadyExists,
		domainrbac.ErrPermissionAlreadyExists,
		domainrbac.ErrRoleAlreadyExists,
		domainrbac.ErrPermissionGroupIDsRequired,
		domainrbac.ErrRoleIDsRequired,
		domainrbac.ErrUnitRequired,
		domainrbac.ErrTargetAccountRequired,
		domainrbac.ErrCannotGrantToOwner,
		domainrbac.ErrPermissionCodeEmpty,
		domainrbac.ErrAccountNotFound,
	}
	for _, known := range knownErrors {
		if errors.Is(err, known) {
			return true
		}
	}
	return false
}
