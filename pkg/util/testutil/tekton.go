package testutil

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func doInstallTektonPipeline(ctx context.Context, cl client.Client, version string) error {
	return doInstall(ctx, cl, "tekton", "tekton-pipelines", version)
}

func InstallTektonPipeline(t *testing.T, ctx context.Context, cl client.Client, version string) {
	require.NoError(t, doInstallTektonPipeline(ctx, cl, version))
}