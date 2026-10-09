package rbac

import (
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
)

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
			err := validate.Struct(UnitAccountRolesCommand{UnitID: 7, AccountID: tt.accountID, RoleIDs: []int{3}})
			if (err == nil) != tt.valid {
				t.Fatalf("validation error = %v, want valid=%v", err, tt.valid)
			}
		})
	}
	if err := validate.Struct(UnitGrantsQuery{UnitID: 7}); err != nil {
		t.Fatalf("omitted account filter must remain valid: %v", err)
	}
}
