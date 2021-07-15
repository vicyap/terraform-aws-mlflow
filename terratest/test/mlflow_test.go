package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/aws"
	http_helper "github.com/gruntwork-io/terratest/modules/http-helper"
	"github.com/gruntwork-io/terratest/modules/terraform"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
)

func TestPublicDeployment(t *testing.T) {
	t.Parallel()

	terraformDir := test_structure.CopyTerraformFolderToTemp(t, "../..", "terratest/examples")

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformDir,
		Vars: map[string]interface{}{
			"is_private": false,
		},
		NoColor: true,
	})

	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	// Getting outputs
	region := "eu-west-1"
	artifactBucketId := terraform.Output(t, terraformOptions, "artifact_bucket_id")
	loadBalancerDns := terraform.Output(t, terraformOptions, "load_balancer_dns_name")

	// The bucket was created
	aws.AssertS3BucketExists(t, region, artifactBucketId)
	aws.AssertS3BucketVersioningExists(t, region, artifactBucketId)

	// MLFlow is healthy
	url := fmt.Sprintf("http://%s/health", loadBalancerDns)
	http_helper.HttpGetWithRetryWithCustomValidation(t, url, nil, 30, 5*time.Second, func(status int, _response string) bool {
		return status == 200
	})
}

func TestPrivateDeploymentWithCustomBucket(t *testing.T) {
	t.Parallel()

	terraformDir := test_structure.CopyTerraformFolderToTemp(t, "../..", "terratest/examples")

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformDir,
		Vars: map[string]interface{}{
			"is_private":         true,
			"artifact_bucket_id": "my-bucket",
		},
		NoColor: true,
	})

	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	// Getting outputs
	artifactBucketId := terraform.Output(t, terraformOptions, "artifact_bucket_id")
	loadBalancerDns := terraform.Output(t, terraformOptions, "load_balancer_dns_name")

	// The bucket was created
	if artifactBucketId != "my-bucket" {
		t.Error("Expected resulting bucket id to be the same we specified")
	}

	url := fmt.Sprintf("http://%s/health", loadBalancerDns)
	err := http_helper.HttpGetWithRetryWithCustomValidationE(t, url, nil, 30, 5*time.Second, func(status int, _response string) bool {
		return status == 200
	})
	if err == nil {
		t.Error("Expected load balancer not to be reachable from the Internet")
	}
}
