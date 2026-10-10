package rbac

import "testing"

func TestBuiltinPermissionsContainsOnlySupportedCodes(t *testing.T) {
	want := []string{
		PermissionUnitView,
		PermissionUnitUpdate,
		PermissionRoomView,
		PermissionRoomManage,
		PermissionRoomTypeView,
		PermissionRentBillView,
		PermissionRentBillManage,
		PermissionMeterReadingView,
		PermissionMeterReadingManage,
		PermissionMeterItemView,
	}
	permissions := BuiltinPermissions()
	if len(permissions) != len(want) {
		t.Fatalf("catalog size = %d, want %d", len(permissions), len(want))
	}
	for i, code := range want {
		if permissions[i].Code != code || permissions[i].Description == "" || !IsBuiltinPermission(code) {
			t.Fatalf("catalog entry %d = %+v, want supported code %q with a description", i, permissions[i], code)
		}
	}
	for _, code := range []string{
		"", "custom.permission", "UNIT.VIEW", " unit.view", "unit.view ",
		PermissionRoomTypeManage, PermissionMeterItemManage, SystemPermissionAllowAllCode,
	} {
		if IsBuiltinPermission(code) {
			t.Errorf("unsupported permission %q accepted", code)
		}
	}
}

func TestBuiltinPermissionsReturnsIndependentCopies(t *testing.T) {
	first, second := BuiltinPermissions(), BuiltinPermissions()
	for i := range first {
		if first[i] == second[i] {
			t.Fatalf("catalog entry %d shares a pointer", i)
		}
	}
	first[0].Code = "custom.permission"
	first[0].Description = "changed"
	first[1] = first[0]
	for i, permission := range BuiltinPermissions() {
		if *permission != *second[i] {
			t.Fatalf("mutating a returned catalog changed entry %d: %+v", i, permission)
		}
	}
}
