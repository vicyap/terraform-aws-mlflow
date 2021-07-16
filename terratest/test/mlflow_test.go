package test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/aws"
	http_helper "github.com/gruntwork-io/terratest/modules/http-helper"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
)

func setupNetworking(t *testing.T, awsRegion string, randomId string) *terraform.Options {
	terraformDir := test_structure.CopyTerraformFolderToTemp(t, "../..", "terratest/examples/networking")

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformDir,
		Vars: map[string]interface{}{
			"random_id": randomId,
		},
		EnvVars: map[string]string{
			"AWS_DEFAULT_REGION": awsRegion,
		},
		NoColor: true,
	})

	terraform.InitAndApply(t, terraformOptions)

	return terraformOptions
}

func setupMlflow(t *testing.T, awsRegion string, networkingTerraformOptions *terraform.Options, vars map[string]interface{}) *terraform.Options {
	terraformDir := test_structure.CopyTerraformFolderToTemp(t, "../..", "terratest/examples/mlflow")

	vars["vpc_id"] = terraform.Output(t, networkingTerraformOptions, "vpc_id")
	vars["vpc_cidr_block"] = terraform.Output(t, networkingTerraformOptions, "vpc_cidr_block")
	vars["private_subnet_ids"] = terraform.OutputList(t, networkingTerraformOptions, "private_subnet_ids")
	vars["public_subnet_ids"] = terraform.OutputList(t, networkingTerraformOptions, "public_subnet_ids")
	vars["database_subnet_ids"] = terraform.OutputList(t, networkingTerraformOptions, "database_subnet_ids")

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformDir,
		Vars:         vars,
		EnvVars: map[string]string{
			"AWS_DEFAULT_REGION": awsRegion,
		},
		NoColor: true,
	})

	terraform.InitAndApply(t, terraformOptions)

	return terraformOptions
}

func assertMlflowPublicDeployment(t *testing.T, terraformOptions *terraform.Options, awsRegion string) {
	// Getting outputs
	artifactBucketId := terraform.Output(t, terraformOptions, "artifact_bucket_id")
	loadBalancerDns := terraform.Output(t, terraformOptions, "load_balancer_dns_name")

	// The bucket was created
	aws.AssertS3BucketExists(t, awsRegion, artifactBucketId)
	aws.AssertS3BucketVersioningExists(t, awsRegion, artifactBucketId)

	// MLFlow is healthy
	url := fmt.Sprintf("http://%s/health", loadBalancerDns)
	http_helper.HttpGetWithRetryWithCustomValidation(t, url, nil, 30, 5*time.Second, func(status int, _response string) bool {
		return status == 200
	})
}

func TestPublicDeploymentEuWest1(t *testing.T) {
	t.Parallel()

	randomId := strings.ToLower(random.UniqueId())
	awsRegion := "eu-west-1"

	networkingTerraformOptions := setupNetworking(t, awsRegion, randomId)
	defer terraform.Destroy(t, networkingTerraformOptions)

	mlflowTerraformOptions := setupMlflow(
		t,
		awsRegion,
		networkingTerraformOptions,
		map[string]interface{}{
			"random_id":  randomId,
			"is_private": false,
		},
	)

	defer terraform.Destroy(t, mlflowTerraformOptions)

	assertMlflowPublicDeployment(t, mlflowTerraformOptions, awsRegion)
}

func TestPublicDeploymentUsWest2(t *testing.T) {
	t.Parallel()

	randomId := strings.ToLower(random.UniqueId())
	awsRegion := "us-west-2"

	networkingTerraformOptions := setupNetworking(t, awsRegion, randomId)
	defer terraform.Destroy(t, networkingTerraformOptions)

	mlflowTerraformOptions := setupMlflow(
		t,
		awsRegion,
		networkingTerraformOptions,
		map[string]interface{}{
			"random_id":  randomId,
			"is_private": false,
		},
	)

	defer terraform.Destroy(t, mlflowTerraformOptions)

	assertMlflowPublicDeployment(t, mlflowTerraformOptions, awsRegion)
}

func TestPrivateDeploymentWithCustomBucket(t *testing.T) {
	t.Parallel()

	randomId := strings.ToLower(random.UniqueId())
	awsRegion := "eu-west-1"

	networkingTerraformOptions := setupNetworking(t, awsRegion, randomId)
	defer terraform.Destroy(t, networkingTerraformOptions)

	mlflowTerraformOptions := setupMlflow(
		t,
		awsRegion,
		networkingTerraformOptions,
		map[string]interface{}{
			"random_id":  randomId,
			"is_private":         true,
			"artifact_bucket_id": "my-bucket",
		},
	)

	defer terraform.Destroy(t, mlflowTerraformOptions)

	// Getting outputs
	artifactBucketId := terraform.Output(t, mlflowTerraformOptions, "artifact_bucket_id")
	loadBalancerDns := terraform.Output(t, mlflowTerraformOptions, "load_balancer_dns_name")

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
