package rbac

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"

	domainrbac "github.com/superwhys/rent-manager/internal/domain/rbac"
)

func TestBindingResponsesPreserveIDsAndPermissionCodes(t *testing.T) {
	role := &domainrbac.Role{ID: 1, PermissionGroupIDs: []int{2, 3}}
	group := &domainrbac.PermissionGroup{ID: 2, PermissionCodes: []string{domainrbac.PermissionRoomView, domainrbac.PermissionRoomManage}}
	roles := ToRoleListResponse([]*domainrbac.Role{role})
	groups := ToPermissionGroupListResponse([]*domainrbac.PermissionGroup{group})
	grant := ToUnitGrantResponse(&domainrbac.UnitAccountRole{Role: role})
	if !reflect.DeepEqual(roles[0].PermissionGroupIDs, []int{2, 3}) ||
		!reflect.DeepEqual(grant.Role.PermissionGroupIDs, []int{2, 3}) ||
		!reflect.DeepEqual(groups[0].PermissionCodes, group.PermissionCodes) {
		t.Fatalf("bindings missing from responses: roles=%+v, groups=%+v, grant=%+v", roles[0], groups[0], grant.Role)
	}
}

func TestBindingResponsesUseEmptyArrays(t *testing.T) {
	for _, test := range []struct {
		name  string
		value any
		key   string
	}{
		{"role", ToRoleResponse(&domainrbac.Role{}), "permission_group_ids"},
		{"permission group", ToPermissionGroupResponse(&domainrbac.PermissionGroup{}), "permission_codes"},
	} {
		t.Run(test.name, func(t *testing.T) {
			data, err := json.Marshal(test.value)
			if err != nil {
				t.Fatal(err)
			}
			var response map[string]json.RawMessage
			if err := json.Unmarshal(data, &response); err != nil {
				t.Fatal(err)
			}
			if string(response[test.key]) != "[]" {
				t.Fatalf("%s = %s, want []", test.key, response[test.key])
			}
		})
	}
}

func TestUnitAccountRolesCommandValidatesAuthkitAccountID(t *testing.T) {
	validate := validator.New()
	tests := []struct {
		name      string
		accountID string
		valid     bool
	}{
		{name: "authkit ID", accountID: strings.Repeat("ab", 32), valid: true},
		{name: "missing"},
		{name: "legacy numeric ID", accountID: "10"},
		{name: "uppercase ID", accountID: strings.Repeat("AB", 32)},
		{name: "non-hex ID", accountID: strings.Repeat("z", 64)},
		{name: "too long", accountID: strings.Repeat("a", 65)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Struct(
				UnitAccountRolesCommand{UnitID: 7, AccountID: tt.accountID, RoleIDs: []int{3}},
			)
			if (err == nil) != tt.valid {
				t.Fatalf("validation error = %v, want valid=%v", err, tt.valid)
			}
		})
	}
	if err := validate.Struct(UnitGrantsQuery{UnitID: 7}); err != nil {
		t.Fatalf("omitted account filter must remain valid: %v", err)
	}
}
