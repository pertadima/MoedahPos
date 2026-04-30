package rbac

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsSuperAdmin(t *testing.T) {
	assert.True(t, IsSuperAdmin("superadmin"))
	assert.False(t, IsSuperAdmin("admin"))
	assert.False(t, IsSuperAdmin("cashier"))
}
