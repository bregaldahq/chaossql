package engine

import (
	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/drivers"
)

// SupportedIsolation returns spec unchanged when driver supports its
// isolation level. Otherwise it clears the level, so the driver runs at its
// own default, and returns the level it dropped.
//
// Only commands that move one spec across engines (diff, matrix, swarm, and
// run with --driver) use it: a level that exists on one engine, such as
// READ_UNCOMMITTED on SQLite, must not turn every other engine into an error.
// A plain run keeps failing on an unsupported level.
func SupportedIsolation(spec domain.Spec, driver drivers.DatabaseDriver) (domain.Spec, domain.IsolationLevel) {
	if spec.Database.Isolation == "" {
		return spec, ""
	}
	if _, err := driver.EffectiveIsolation(spec.Database.Isolation); err == nil {
		return spec, ""
	}
	dropped := spec.Database.Isolation
	spec.Database.Isolation = ""
	return spec, dropped
}
