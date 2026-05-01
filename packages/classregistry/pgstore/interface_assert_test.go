package pgstore

import (
	"p9e.in/samavaya/packages/classregistry"
)

// Compile-time assertion: *Store satisfies classregistry.OverrideStore.
// Breaking this contract (renaming ListForTenantDomain, changing its
// signature, etc.) fails the build before any caller sees a runtime
// error.
var _ classregistry.OverrideStore = (*Store)(nil)
