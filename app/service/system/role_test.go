package system

import (
	"testing"

	systemModel "github.com/seakee/go-api/app/model/system"
	"gorm.io/gorm"
)

func TestBuildRoleMenuPermissionTree(t *testing.T) {
	menus := systemModel.MenuList{
		{
			Model:        gorm.Model{ID: 1},
			Name:         "系统管理",
			Path:         "/system",
			PermissionId: 0,
			ParentId:     0,
			Children: systemModel.MenuList{
				{
					Model:        gorm.Model{ID: 2},
					Name:         "用户管理",
					Path:         "/system/user",
					PermissionId: 101,
					ParentId:     1,
				},
				{
					Model:        gorm.Model{ID: 3},
					Name:         "角色管理",
					Path:         "/system/role",
					PermissionId: 102,
					ParentId:     1,
				},
			},
		},
		{
			Model:        gorm.Model{ID: 4},
			Name:         "审计日志",
			Path:         "/audit",
			PermissionId: 201,
			ParentId:     0,
		},
	}

	permissionSet := map[uint]struct{}{
		101: {},
		201: {},
	}

	got := buildRoleMenuPermissionTree(menus, permissionSet)

	if len(got) != 2 {
		t.Fatalf("buildRoleMenuPermissionTree() root len = %d, want 2", len(got))
	}

	if got[0].Checked {
		t.Fatalf("root menu checked = true, want false")
	}

	if len(got[0].Children) != 2 {
		t.Fatalf("first root children len = %d, want 2", len(got[0].Children))
	}

	if !got[0].Children[0].Checked {
		t.Fatalf("user menu checked = false, want true")
	}

	if got[0].Children[1].Checked {
		t.Fatalf("role menu checked = true, want false")
	}

	if !got[1].Checked {
		t.Fatalf("audit menu checked = false, want true")
	}
}
