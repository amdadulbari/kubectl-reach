// Copyright 2025 The kubectl-reach Authors.
// Licensed under the Apache License, Version 2.0.

package reach

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/kubectl-reach/kubectl-reach/pkg/version"
)

// NewCommand returns the reach cobra command (kubectl reach).
func NewCommand(streams genericclioptions.IOStreams) *cobra.Command {
	o := &Options{IOStreams: streams}
	cmd := &cobra.Command{
		Use:   "reach <source-pod-name> --to <target:port> [flags]",
		Short: "Test network connectivity from a pod to a target using an ephemeral debug container",
		Long: `Test network connectivity from an existing Pod to a target (IP, DNS, or Service)
by injecting an ephemeral debug container into the source Pod. Useful for verifying
if NetworkPolicies or Service Meshes are blocking traffic.

Requires Kubernetes 1.25+ (EphemeralContainers is stable). Does not restart existing
application containers.`,
		Example: `  kubectl reach myapp-abc123 --to google.com:443
  kubectl reach myapp-abc123 --to 10.0.0.5:8080 -n myns
  kubectl reach myapp-abc123 --to myservice:80 --image busybox --timeout 10`,
		Args:    cobra.ExactArgs(1),
		Version: version.Version,
		RunE: func(cmd *cobra.Command, args []string) error {
			o.PodName = args[0]
			return o.Run(cmd.Context())
		},
	}

	configFlags := genericclioptions.NewConfigFlags(true)
	configFlags.AddFlags(cmd.PersistentFlags())
	o.ConfigFlags = configFlags

	cmd.Flags().StringVar(&o.To, "to", "", "Target destination as host:port (e.g. google.com:443 or 10.0.0.5:8080)")
	cmd.Flags().StringVar(&o.Image, "image", DefaultImage, "Debug container image (default: busybox)")
	cmd.Flags().DurationVar(&o.Timeout, "timeout", DefaultTimeout*time.Second, "Timeout for the connection check")

	_ = cmd.MarkFlagRequired("to")

	usageWithKubectlPrefix(cmd)
	return cmd
}

func usageWithKubectlPrefix(cmd *cobra.Command) {
	base := filepath.Base(os.Args[0])
	if strings.HasPrefix(base, "kubectl-") {
		cmd.SetUsageTemplate(strings.ReplaceAll(cmd.UsageTemplate(), "Usage:", "Usage:\n  kubectl "+strings.TrimPrefix(base, "kubectl-")+" "))
	}
}
