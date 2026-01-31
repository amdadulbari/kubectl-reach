// Copyright 2025 The kubectl-reach Authors.
// Licensed under the Apache License, Version 2.0.

package reach

import (
	"context"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
)

const (
	// EphemeralContainerNamePrefix is the prefix for ephemeral container names.
	// A unique suffix (timestamp) is added per run so you can run reach multiple times on the same pod.
	EphemeralContainerNamePrefix = "reach-debug"
	EphemeralContainerName       = EphemeralContainerNamePrefix // kept for backward compatibility / tests
	DefaultImage                 = "busybox"
	DefaultTimeout               = 5
)

// Options holds the reach command options and runtime.
type Options struct {
	genericclioptions.IOStreams
	ConfigFlags *genericclioptions.ConfigFlags
	PodName     string
	To          string
	Image       string
	Timeout     time.Duration
}

// Run runs the reach logic: add ephemeral container, wait, stream logs.
func (o *Options) Run(ctx context.Context) error {
	host, port, err := parseTarget(o.To)
	if err != nil {
		return fmt.Errorf("invalid --to %q: %w", o.To, err)
	}

	restConfig, err := o.ConfigFlags.ToRESTConfig()
	if err != nil {
		return fmt.Errorf("building rest config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("creating kubernetes client: %w", err)
	}

	namespace, _, err := o.ConfigFlags.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return fmt.Errorf("getting namespace: %w", err)
	}
	if namespace == "" {
		namespace = corev1.NamespaceDefault
	}

	pod, err := clientset.CoreV1().Pods(namespace).Get(ctx, o.PodName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return fmt.Errorf("pod %q not found in namespace %q", o.PodName, namespace)
		}
		return fmt.Errorf("getting pod: %w", err)
	}

	timeoutSec := int(o.Timeout.Seconds())
	if timeoutSec < 1 {
		timeoutSec = 1
	}

	// Unique name per run so you can run reach multiple times on the same pod.
	ephemeralName := EphemeralContainerNamePrefix + "-" + strconv.FormatInt(time.Now().UnixNano(), 10)

	ec := corev1.EphemeralContainer{
		EphemeralContainerCommon: corev1.EphemeralContainerCommon{
			Name:    ephemeralName,
			Image:   o.Image,
			Command: []string{"nc"},
			Args:    []string{"-zv", "-w", strconv.Itoa(timeoutSec), host, port},
		},
	}

	podCopy := pod.DeepCopy()
	if podCopy.Spec.EphemeralContainers == nil {
		podCopy.Spec.EphemeralContainers = []corev1.EphemeralContainer{}
	}
	podCopy.Spec.EphemeralContainers = append(podCopy.Spec.EphemeralContainers, ec)

	_, err = clientset.CoreV1().Pods(namespace).UpdateEphemeralContainers(ctx, o.PodName, podCopy, metav1.UpdateOptions{})
	if err != nil {
		if isEphemeralContainersUnsupported(err) {
			return fmt.Errorf("ephemeral containers are not supported by this cluster (requires Kubernetes 1.25+): %w", err)
		}
		return fmt.Errorf("adding ephemeral container: %w", err)
	}

	if err := o.waitForEphemeralContainer(ctx, clientset, namespace, o.PodName, ephemeralName); err != nil {
		return err
	}

	return o.streamLogs(ctx, clientset, namespace, o.PodName, ephemeralName)
}

func parseTarget(to string) (host, port string, err error) {
	host, port, err = net.SplitHostPort(to)
	if err != nil {
		return "", "", fmt.Errorf("expected host:port: %w", err)
	}
	if host == "" || port == "" {
		return "", "", fmt.Errorf("host and port must be non-empty")
	}
	return host, port, nil
}

func isEphemeralContainersUnsupported(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "ephemeralcontainers") ||
		strings.Contains(s, "EphemeralContainers") ||
		strings.Contains(s, "subresource") ||
		strings.Contains(s, "MethodNotAllowed") ||
		errors.IsNotFound(err)
}

func (o *Options) waitForEphemeralContainer(ctx context.Context, clientset kubernetes.Interface, namespace, podName, containerName string) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	err := wait.PollUntilContextCancel(timeoutCtx, 500*time.Millisecond, true, func(ctx context.Context) (bool, error) {
		pod, err := clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		for _, s := range pod.Status.EphemeralContainerStatuses {
			if s.Name == containerName {
				if s.State.Running != nil {
					return true, nil
				}
				if s.State.Terminated != nil {
					return true, nil
				}
			}
		}
		return false, nil
	})
	if err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("timeout waiting for ephemeral container to start: %w", err)
		}
		return fmt.Errorf("waiting for ephemeral container: %w", err)
	}
	return nil
}

func (o *Options) streamLogs(ctx context.Context, clientset kubernetes.Interface, namespace, podName, containerName string) error {
	opts := &corev1.PodLogOptions{
		Container:  containerName,
		Follow:     true,
		Previous:   false,
		Timestamps: false,
	}

	req := clientset.CoreV1().Pods(namespace).GetLogs(podName, opts)
	stream, err := req.Stream(ctx)
	if err != nil {
		return fmt.Errorf("streaming logs from ephemeral container: %w", err)
	}
	defer stream.Close()

	_, err = io.Copy(o.Out, stream)
	if err != nil && err != io.EOF {
		return fmt.Errorf("copying logs: %w", err)
	}
	return nil
}
