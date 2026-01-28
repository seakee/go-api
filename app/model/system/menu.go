// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package system

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"gorm.io/gorm"
)

// Menu represents a menu item in the system menu hierarchy.
type Menu struct {
	gorm.Model

	Name         string   `gorm:"column:name" json:"name"`
	Path         string   `gorm:"column:path" json:"path"`
	PermissionId uint     `gorm:"column:permission_id" json:"permission_id"`
	ParentId     uint     `gorm:"column:parent_id" json:"parent_id"`
	Icon         string   `gorm:"column:icon" json:"icon"`
	Sort         int      `gorm:"column:sort" json:"sort"`
	Children     MenuList `json:"children,omitempty" gorm:"-"`

	// Query condition fields for Where method
	queryCondition string        `gorm:"-" json:"-"`
	queryArgs      []interface{} `gorm:"-" json:"-"`
}

// MenuList represents a list of Menu items that implements sort.Interface.
type MenuList []Menu

func (l MenuList) Len() int           { return len(l) }
func (l MenuList) Less(i, j int) bool { return l[i].Sort < l[j].Sort }
func (l MenuList) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }

// getChildren retrieves child menus for a given parent ID.
func (l MenuList) getChildren(id uint, treeMap map[uint]MenuList) MenuList {
	children := treeMap[id]
	if len(children) == 0 {
		return MenuList{}
	}

	for i := range children {
		children[i].Children = l.getChildren(children[i].ID, treeMap)
	}
	sort.Sort(children)
	return children
}

// GenTree generates a tree structure from the flat menu list.
func (l MenuList) GenTree() MenuList {
	treeMap := make(map[uint]MenuList, len(l))
	for _, v := range l {
		treeMap[v.ParentId] = append(treeMap[v.ParentId], v)
	}

	tree := treeMap[0]
	for i := range tree {
		tree[i].Children = l.getChildren(tree[i].ID, treeMap)
	}
	sort.Sort(tree)
	return tree
}

// AllUserMenuIds finds all menu IDs within the user's permission scope.
func (l MenuList) AllUserMenuIds(menuId uint, menuIdsMap map[uint]struct{}) {
	for _, menu := range l {
		if menu.ID != menuId {
			continue
		}
		if menu.ParentId == 0 {
			menuIdsMap[menu.ID] = struct{}{}
			return
		}
		menuIdsMap[menu.ID] = struct{}{}
		l.AllUserMenuIds(menu.ParentId, menuIdsMap)
		return
	}
}

// TableName specifies the table name for the Menu model.
func (m *Menu) TableName() string {
	return "sys_menu"
}

// Where sets query conditions for the Menu model.
//
// Parameters:
//   - condition: string SQL condition with placeholders.
//   - args: ...interface{} arguments to replace placeholders in the condition.
//
// Returns:
//   - *Menu: pointer to the Menu instance for method chaining.
func (m *Menu) Where(condition string, args ...interface{}) *Menu {
	m.queryCondition = condition
	m.queryArgs = args
	return m
}

// First retrieves the first menu matching the criteria from the database.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//
// Returns:
//   - *Menu: pointer to the retrieved menu, or nil if not found.
//   - error: error if the query fails, otherwise nil.
func (m *Menu) First(ctx context.Context, db *gorm.DB) (*Menu, error) {
	var menu Menu

	// Build query with context
	query := db.WithContext(ctx)

	// Apply Where condition if set, otherwise use struct fields
	if m.queryCondition != "" {
		query = query.Where(m.queryCondition, m.queryArgs...)
	} else {
		query = query.Where(m)
	}

	// Perform the database query
	if err := query.First(&menu).Error; err != nil {
		// If no record is found, return nil without an error.
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		// Return the error if the query fails.
		return nil, fmt.Errorf("find first failed: %w", err)
	}

	return &menu, nil
}

// Last retrieves the last menu matching the criteria from the database, ordered by ID in descending order.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//
// Returns:
//   - *Menu: pointer to the retrieved menu, or nil if not found.
//   - error: error if the query fails, otherwise nil.
func (m *Menu) Last(ctx context.Context, db *gorm.DB) (*Menu, error) {
	var menu Menu

	// Build query with context
	query := db.WithContext(ctx)

	// Apply Where condition if set, otherwise use struct fields
	if m.queryCondition != "" {
		query = query.Where(m.queryCondition, m.queryArgs...)
	} else {
		query = query.Where(m)
	}

	// Perform the database query
	if err := query.Order("id desc").First(&menu).Error; err != nil {
		// If no record is found, return nil without an error.
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		// Return the error if the query fails.
		return nil, fmt.Errorf("find last failed: %w", err)
	}

	return &menu, nil
}

// Create inserts a new menu into the database and returns the ID of the created Menu.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//
// Returns:
//   - uint: ID of the created menu.
//   - error: error if the insert operation fails, otherwise nil.
func (m *Menu) Create(ctx context.Context, db *gorm.DB) (uint, error) {
	// Perform the database insert operation with context.
	if err := db.WithContext(ctx).Create(m).Error; err != nil {
		return 0, fmt.Errorf("create failed: %w", err)
	}

	return m.ID, nil
}

// Delete removes the menu from the database.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//
// Returns:
//   - error: error if the delete operation fails, otherwise nil.
func (m *Menu) Delete(ctx context.Context, db *gorm.DB) error {
	// Build query with context
	query := db.WithContext(ctx)

	// Apply Where condition if set, otherwise use struct fields
	if m.queryCondition != "" {
		return query.Where(m.queryCondition, m.queryArgs...).Delete(&Menu{}).Error
	} else {
		return query.Where(m).Delete(&Menu{}).Error
	}
}

// Updates applies the specified updates to the menu in the database.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//   - updates: map[string]interface{} containing the updates to apply.
//
// Returns:
//   - error: error if the update operation fails, otherwise nil.
func (m *Menu) Updates(ctx context.Context, db *gorm.DB, updates map[string]interface{}) error {
	// Build query with context and explicitly set the model
	query := db.WithContext(ctx).Model(&Menu{})

	// Apply Where condition if set, otherwise use struct fields
	if m.queryCondition != "" {
		return query.Where(m.queryCondition, m.queryArgs...).Updates(updates).Error
	} else {
		return query.Where(m).Updates(updates).Error
	}
}

// List retrieves all menus matching the criteria from the database.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//
// Returns:
//   - []Menu: slice of retrieved menus.
//   - error: error if the query fails, otherwise nil.
func (m *Menu) List(ctx context.Context, db *gorm.DB) ([]Menu, error) {
	var menus []Menu

	// Build query with context
	query := db.WithContext(ctx)

	// Apply Where condition if set, otherwise use struct fields
	if m.queryCondition != "" {
		query = query.Where(m.queryCondition, m.queryArgs...)
	} else {
		query = query.Where(m)
	}

	// Perform the database query
	if err := query.Find(&menus).Error; err != nil {
		return nil, fmt.Errorf("list failed: %w", err)
	}

	return menus, nil
}

// ListByArgs retrieves menus matching the specified query and arguments from the database, ordered by ID in descending order.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//   - query: SQL query string.
//   - args: variadic arguments for the SQL query.
//
// Returns:
//   - []Menu: slice of retrieved menus.
//   - error: error if the query fails, otherwise nil.
func (m *Menu) ListByArgs(ctx context.Context, db *gorm.DB, query interface{}, args ...interface{}) ([]Menu, error) {
	var menus []Menu

	// Perform the database query with context.
	if err := db.WithContext(ctx).Where(query, args...).Order("id desc").Find(&menus).Error; err != nil {
		return nil, fmt.Errorf("list by args failed: %w", err)
	}

	return menus, nil
}

// CountByArgs counts the number of menus matching the specified query and arguments in the database.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//   - query: SQL query string.
//   - args: variadic arguments for the SQL query.
//
// Returns:
//   - int64: count of matching menus.
//   - error: error if the count operation fails, otherwise nil.
func (m *Menu) CountByArgs(ctx context.Context, db *gorm.DB, query interface{}, args ...interface{}) (int64, error) {
	var count int64

	// Perform the database count operation with context.
	if err := db.WithContext(ctx).Model(&Menu{}).Where(query, args...).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count by args failed: %w", err)
	}

	return count, nil
}

// Count counts the number of menus matching the criteria in the database.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//
// Returns:
//   - int64: count of matching menus.
//   - error: error if the count operation fails, otherwise nil.
func (m *Menu) Count(ctx context.Context, db *gorm.DB) (int64, error) {
	var count int64

	// Build query with context
	query := db.WithContext(ctx).Model(&Menu{})

	// Apply Where condition if set, otherwise use struct fields
	if m.queryCondition != "" {
		query = query.Where(m.queryCondition, m.queryArgs...)
	} else {
		query = query.Where(m)
	}

	// Perform the database count operation
	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count failed: %w", err)
	}

	return count, nil
}

// BatchInsert inserts multiple menus into the database in a single batch operation.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//   - menus: slice of Menu instances to be inserted.
//
// Returns:
//   - error: error if the batch insert operation fails, otherwise nil.
func (m *Menu) BatchInsert(ctx context.Context, db *gorm.DB, menus []Menu) error {
	// Perform the database batch insert operation with context.
	return db.WithContext(ctx).Create(&menus).Error
}

// FindWithPagination retrieves menus matching the criteria from the database with pagination support.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//   - page: page number for pagination (1-based).
//   - size: number of menus per page.
//
// Returns:
//   - []Menu: slice of retrieved menus.
//   - error: error if the query fails, otherwise nil.
func (m *Menu) FindWithPagination(ctx context.Context, db *gorm.DB, page, size int) ([]Menu, error) {
	var menus []Menu

	// Build query with context
	query := db.WithContext(ctx)

	// Apply Where condition if set, otherwise use struct fields
	if m.queryCondition != "" {
		query = query.Where(m.queryCondition, m.queryArgs...)
	} else {
		query = query.Where(m)
	}

	// Perform the database query with context, applying offset and limit for pagination.
	if err := query.Offset((page - 1) * size).Limit(size).Order("id desc").Find(&menus).Error; err != nil {
		// Return the error if the query fails.
		return nil, fmt.Errorf("find with pagination failed: %w", err)
	}

	return menus, nil
}

// FindWithSort retrieves menus matching the criteria from the database with sorting support.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//   - sort: sorting criteria (e.g., "id desc").
//
// Returns:
//   - []Menu: slice of retrieved menus.
//   - error: error if the query fails, otherwise nil.
func (m *Menu) FindWithSort(ctx context.Context, db *gorm.DB, sort string) ([]Menu, error) {
	var menus []Menu

	// Build query with context
	query := db.WithContext(ctx)

	// Apply Where condition if set, otherwise use struct fields
	if m.queryCondition != "" {
		query = query.Where(m.queryCondition, m.queryArgs...)
	} else {
		query = query.Where(m)
	}

	// Perform the database query with context, applying the specified sort order.
	if err := query.Order(sort).Find(&menus).Error; err != nil {
		// Return the error if the query fails.
		return nil, fmt.Errorf("find with sort failed: %w", err)
	}

	return menus, nil
}
