// Copyright 2025 The kubectl-reach Authors.
// Licensed under the Apache License, Version 2.0.

package reach

import (
	"errors"
	"os"
	"testing"

	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func TestParseTarget(t *testing.T) {
	tests := []struct {
		name      string
		to        string
		wantHost  string
		wantPort  string
		wantError bool
	}{
		{"valid host:port", "google.com:443", "google.com", "443", false},
		{"valid IP:port", "10.0.0.5:8080", "10.0.0.5", "8080", false},
		{"valid host:port with high port", "myservice:80", "myservice", "80", false},
		{"localhost", "localhost:9090", "localhost", "9090", false},
		{"empty", "", "", "", true},
		{"missing port", "hostonly", "", "", true},
		{"missing host", ":443", "", "", true},
		{"bad format", "host:port:extra", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host, port, err := parseTarget(tt.to)
			if (err != nil) != tt.wantError {
				t.Errorf("parseTarget(%q) error = %v, wantError %v", tt.to, err, tt.wantError)
				return
			}
			if !tt.wantError {
				if host != tt.wantHost || port != tt.wantPort {
					t.Errorf("parseTarget(%q) = %q, %q; want %q, %q", tt.to, host, port, tt.wantHost, tt.wantPort)
				}
			}
		})
	}
}

func TestIsEphemeralContainersUnsupported(t *testing.T) {
	// errors.IsNotFound requires k8s types; we test string-based cases here.

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"generic error", errors.New("something failed"), false},
		{"ephemeralcontainers lowercase", errors.New("the server could not find the requested resource (patch pods ephemeralcontainers)"), true},
		{"EphemeralContainers mixed", errors.New("EphemeralContainers not supported"), true},
		{"subresource", errors.New("subresource ephemeralcontainers is not supported"), true},
		{"MethodNotAllowed", errors.New("405 Method Not Allowed: ephemeralcontainers"), true},
		{"unrelated", errors.New("pod not found"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isEphemeralContainersUnsupported(tt.err)
			if got != tt.want {
				t.Errorf("isEphemeralContainersUnsupported(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestConstants(t *testing.T) {
	if EphemeralContainerName != "reach-debug" {
		t.Errorf("EphemeralContainerName = %q, want reach-debug", EphemeralContainerName)
	}
	if DefaultImage != "busybox" {
		t.Errorf("DefaultImage = %q, want busybox", DefaultImage)
	}
	if DefaultTimeout != 5 {
		t.Errorf("DefaultTimeout = %d, want 5", DefaultTimeout)
	}
}

func TestNewCommand(t *testing.T) {
	streams := genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
	cmd := NewCommand(streams)
	if cmd == nil {
		t.Fatal("NewCommand() returned nil")
	}
	if cmd.Use != "reach <source-pod-name> --to <target:port> [flags]" {
		t.Errorf("Use = %q", cmd.Use)
	}
	// Required flag --to
	f := cmd.Flags().Lookup("to")
	if f == nil {
		t.Fatal("flag --to not found")
	}
	if f.Usage == "" {
		t.Error("flag --to has empty usage")
	}
	// Run with --help should not error
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("Execute() with --help: %v", err)
	}
}
