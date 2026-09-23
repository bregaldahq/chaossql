package cloud

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/bregaldahq/chaossql/internal/domain"
)

// ScenarioFingerprint hashes the resolved scenario semantics locally. SQL and
// parameter values never leave this function; connection credentials, prose and
// execution PRNG seed are excluded. The database seed SQL remains part of the
// scenario identity. JSON sorts parameter map keys for deterministic serialization.
func ScenarioFingerprint(spec domain.Spec) (string, error) {
	spec.Database.DSN = ""
	spec.Description = ""
	spec.Engine.Seed = 0
	encoded, err := json.Marshal(spec)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(encoded)
	return hex.EncodeToString(hash[:]), nil
}
