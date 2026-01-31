// Copyright 2025 The kubectl-reach Authors.
// Licensed under the Apache License, Version 2.0.

package version

import (
	"testing"
)

func TestVersionNonEmpty(t *testing.T) {
	if Version == "" {
		t.Error("Version should not be empty")
	}
}

func TestVersionFormat(t *testing.T) {
	// Default is "dev"; when built with ldflags it can be "v1.2.3"
	// Just ensure it's a non-empty string
	if len(Version) == 0 {
		t.Error("Version must be set")
	}
}
