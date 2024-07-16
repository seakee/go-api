// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package codegen

import (
	"testing"
)

func TestModel_Generate(t *testing.T) {
	m := NewModel()

	err := m.Generate(false, "go-api/bin/data/sql/jd_account.sql", "")
	if err != nil {
		t.Fatal(err)
	}
}
