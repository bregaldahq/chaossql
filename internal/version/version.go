// Package version is the single source of the ChaosSQL release version.
// Bump it in the release pull request; binaries, reports, and the WASM engine
// read it from here instead of carrying their own copies.
package version

const Version = "1.6.0"
