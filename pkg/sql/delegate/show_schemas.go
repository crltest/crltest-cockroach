// Copyright 2019 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package delegate

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/sql/lexbase"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/cat"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
)

// delegateShowSchemas implements SHOW SCHEMAS which returns all the schemas in
// the given or current database.
// Privileges: None.
func (d *delegator) delegateShowSchemas(n *tree.ShowSchemas) (tree.Statement, error) {
	name, err := d.getSpecifiedOrCurrentDatabase(n.Database)
	if err != nil {
		return nil, err
	}
	getSchemasQuery := fmt.Sprintf(`
      SELECT nspname AS schema_name, rolname AS owner
      FROM %[1]s.information_schema.schemata i
      INNER JOIN %[1]s.pg_catalog.pg_namespace n ON (n.nspname = i.schema_name)
      LEFT JOIN %[1]s.pg_catalog.pg_roles r ON (n.nspowner = r.oid)
			WHERE catalog_name = %[2]s
			ORDER BY schema_name`,
		name.String(), // note: (tree.Name).String() != string(name)
		lexbase.EscapeSQLString(string(name)),
	)

	return d.parse(getSchemasQuery)
}

func (d *delegator) delegateShowCreateAllSchemas() (tree.Statement, error) {
	sqltelemetry.IncrementShowCounter(sqltelemetry.Create)

	const showCreateAllSchemasQuery = `
	SELECT crdb_internal.show_create_all_schemas(%[1]s) AS create_statement;
`
	databaseLiteral := d.evalCtx.SessionData().Database

	query := fmt.Sprintf(showCreateAllSchemasQuery,
		lexbase.EscapeSQLString(databaseLiteral),
	)

	return d.parse(query)
}

// getSpecifiedOrCurrentDatabase returns the name of the specified database, or
// of the current database if the specified name is empty.
//
// Returns an error if there is no current database, or if the specified
// database doesn't exist.
func (d *delegator) getSpecifiedOrCurrentDatabase(specifiedDB tree.Name) (tree.Name, error) {
	flags := cat.Flags{AvoidDescriptorCaches: true}
	return d.catalog.LookupDatabaseName(d.ctx, flags, string(specifiedDB))
}
