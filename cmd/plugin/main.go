// Copyright 2025 The kubectl-reach Authors.
// Licensed under the Apache License, Version 2.0.

package main

import (
	"os"

	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/kubectl-reach/kubectl-reach/pkg/reach"

	// Import auth plugins for cloud provider authentication (GCP, Azure, OIDC, etc.)
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func main() {
	streams := genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
	cmd := reach.NewCommand(streams)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
