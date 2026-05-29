package services

type AdminServices struct {
	Account   *AccountService
	Home      *HomeService
	RBAC      *RBACService
	Unit      *UnitService
	Room      *RoomService
	RoomType  *RoomTypeService
	MeterItem *MeterItemService
}
