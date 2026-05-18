package consts

const (
	CalcTypeDifference = "DIFFERENCE" // 差值型
	CalcTypeDirect     = "DIRECT"     // 直接录入型

	SourceManual = "MANUAL" // 手动录入
	SourceImport = "IMPORT" // 导入
	SourceAuto   = "AUTO"   // 自动计算

	RentBillStatusUnpaid    = "UNPAID"    // 待支付
	RentBillStatusPaid      = "PAID"      // 已支付
	RentBillStatusCancelled = "CANCELLED" // 已取消

	SystemRoleSuperAdmin                = "SuperAdmin"
	SystemPermissionAllowAllCode        = "SYSTEM_ALLOW_ALL"
	SystemPermissionAllowAllDescription = "Allow all APIs"
)
