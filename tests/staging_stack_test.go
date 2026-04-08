package tests

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terragrunt"
)

func TestStaging(t *testing.T) {
	t.Parallel()

	stackDir := "../live/staging/"

	options := &terragrunt.Options{
		// Run from the examples subfolder where the terragrunt configs are
        TerragruntDir: stackDir,
		// Optional: Set log level for cleaner output
		TerragruntArgs: []string{"--log-level", "error"},
	}

	// Clean up all modules with "terragrunt destroy --all" at the end of the test.
	// DestroyAll respects the reverse dependency order.
	defer terragrunt.DestroyAll(t, options)

	// Run "terragrunt apply --all". This applies all modules in dependency order.
	terragrunt.ApplyAll(t, options)

    // Add additional Go tests here if necessary
}
