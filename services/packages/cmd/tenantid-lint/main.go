// Command tenantid-lint scans services/**/migrations/*.sql for
// CREATE TABLE statements that omit `tenant_id`. Run from the CI
// workflow on every PR; exits non-zero when any violation is
// found.
//
// Usage:
//
//	tenantid-lint [root]
//
// `root` defaults to the current directory. The tool walks the
// tree looking for *.sql files inside any `migrations/`
// directory; tables in the lint.Exempt allowlist are skipped.
package main

import (
	"fmt"
	"os"

	"p9e.in/chetana/packages/db/lint"
)

func main() {
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}
	violations, err := lint.CheckMigrations(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tenantid-lint: %v\n", err)
		os.Exit(2)
	}
	for _, v := range violations {
		fmt.Println(v.String())
	}
	if len(violations) > 0 {
		fmt.Fprintf(os.Stderr, "\n%d table(s) missing tenant_id (or not in lint.Exempt).\n",
			len(violations))
		fmt.Fprintln(os.Stderr,
			"Add tenant_id to the table OR (for cross-tenant rows) extend Exempt with reviewer approval.")
		os.Exit(1)
	}
	fmt.Println("OK: every domain migration carries tenant_id.")
}
