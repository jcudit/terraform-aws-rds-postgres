package test

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/terraform"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/stretchr/testify/assert"

	aws_sdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/rds"

	_ "github.com/lib/pq"
)

// Test the Terraform module in examples/stg-us-west-2,
// which is an example use of the root module
func TestStgUswest2(t *testing.T) {
	t.Parallel()

	// Create state for passing data between test stages
	// https://github.com/gruntwork-io/terratest#iterating-locally-using-test-stages
	exampleFolder := test_structure.CopyTerraformFolderToTemp(
		t,
		"../../",
		"examples/stg-us-west-2",
	)

	// At the end of the test, `terraform destroy` the created resources
	defer test_structure.RunTestStage(t, "teardown", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, exampleFolder)
		terraform.Destroy(t, terraformOptions)
	})

	// Deploy the tested infrastructure
	test_structure.RunTestStage(t, "setup", func() {
		terraformOptions := configureTerraformOptions(t, exampleFolder)

		// Save the options and key pair so later test stages can use them
		test_structure.SaveTerraformOptions(t, exampleFolder, terraformOptions)

		// Run `terraform init` and `terraform apply` and fail if there are errors
		terraform.InitAndApply(t, terraformOptions)
	})

	// Validate the test infrastructure
	test_structure.RunTestStage(t, "validate", func() {
		testDatabaseValid(t, exampleFolder)
		testDatabaseSnapshotRestore(t, exampleFolder)
	})
}

func configureTerraformOptions(t *testing.T, exampleFolder string) *terraform.Options {

	terraformOptions := &terraform.Options{
		// The path to where our Terraform code is located
		TerraformDir: exampleFolder,

		// Variables to pass to our Terraform code using -var options
		Vars: map[string]interface{}{
			"environment": "staging",
			"region":      "us-west-2",
		},

		// Environment variables to set when running Terraform
		EnvVars: map[string]string{
			"AWS_DEFAULT_REGION": "us-west-2",
		},
	}

	return terraformOptions
}

func testDatabaseValid(t *testing.T, exampleFolder string) {

	// Configure output validation
	terraformOptions := configureTerraformOptions(t, exampleFolder)

	// The database has valid characteristics
	endpoints := terraform.Output(t, terraformOptions, "aws_rds_cluster_instance_endpoints")
	assert.NotEmpty(t, endpoints)

	// The database is available in multiple AZs
	region := "us-west-2"
	maxRetries := 10
	timeBetweenRetries := 1 * time.Second
	description := fmt.Sprintf("Awaiting creation of database")
	availabilityZones := map[string]bool{}
	instanceIDs := terraform.OutputList(t, terraformOptions, "aws_rds_cluster_instance_ids")
	for _, id := range instanceIDs {
		retry.DoWithRetry(t, description, maxRetries, timeBetweenRetries, func() (string, error) {
			instance, err := aws.GetRdsInstanceDetailsE(t, id, region)
			if err != nil {
				return "", fmt.Errorf("Expected database %s to be present in %s", id, region)
			}

			availabilityZones[*instance.AvailabilityZone] = true
			return "", nil
		})
	}
	assert.Greater(t, len(availabilityZones), 1)

	// The database is reachable
	connectionString := terraform.Output(t, terraformOptions, "reader_connection_string")
	db, err := sql.Open("postgres", connectionString)
	assert.NoError(t, err)
	err = db.Ping()
	assert.NoError(t, err)
}

func testDatabaseSnapshotRestore(t *testing.T, exampleFolder string) {

	// Run this test as needed
	if _, exists := os.LookupEnv("TEST_DATABASE_SNAPSHOT_RESTORE"); !exists {
		t.Skip()
	}

	// Retrieve outputs for validation
	terraformOptions := configureTerraformOptions(t, exampleFolder)
	id := terraform.Output(t, terraformOptions, "aws_rds_cluster_identifier")
	dbSubnetGroupName := terraform.Output(t, terraformOptions, "aws_rds_cluster_db_subnet_group_name")
	dbSecurityGroups := terraform.OutputList(t, terraformOptions, "aws_rds_cluster_security_groups")

	// Setup for snapshot / restore steps below
	now := time.Now()
	entropy := now.Format("20060102150405")
	snapshotID := id + "-" + entropy
	maxRetries := 10
	timeBetweenRetries := 10 * time.Second

	// Obtain API access for RDS operations below
	session, err := aws.NewAuthenticatedSessionFromDefaultCredentials("us-west-2")
	assert.NoError(t, err)
	service := rds.New(session)

	// Create the RDS Cluster snapshot
	snapshot, err := service.CreateDBClusterSnapshot(
		&rds.CreateDBClusterSnapshotInput{
			DBClusterIdentifier:         aws_sdk.String(id),
			DBClusterSnapshotIdentifier: aws_sdk.String(snapshotID),
		},
	)
	assert.NoError(t, err)
	fmt.Printf("[DEBUG] Snapshot Requested: %v\n", snapshot)

	// Wait for snapshot to complete
	description := fmt.Sprintf("Awaiting creation of snapshot")
	snapshotInput := &rds.DescribeDBClusterSnapshotsInput{
		DBClusterSnapshotIdentifier: aws_sdk.String(snapshotID),
		SnapshotType:                aws_sdk.String("manual"),
	}
	snapshotCreated := false
	retry.DoWithRetry(t, description, maxRetries, timeBetweenRetries, func() (string, error) {
		result, err := service.DescribeDBClusterSnapshots(snapshotInput)
		if err != nil {
			return "", err
		}
		json := fmt.Sprintf("%v", result)
		if !strings.Contains(json, `Status: "available"`) {
			return fmt.Sprintf("%v", result), fmt.Errorf("Still creating")
		}
		snapshotCreated = true
		return fmt.Sprintf("%v", result), nil
	})
	assert.True(t, snapshotCreated)

	// Rename existing cluster
	modifyInput := &rds.ModifyDBClusterInput{
		ApplyImmediately:       aws_sdk.Bool(true),
		DBClusterIdentifier:    aws_sdk.String(id),
		NewDBClusterIdentifier: aws_sdk.String(snapshotID),
	}
	modifyResult, err := service.ModifyDBCluster(modifyInput)
	assert.NoError(t, err)
	fmt.Printf("[DEBUG] Cluster Rename Requested: %v\n", modifyResult)

	// Wait for rename to complete
	description = fmt.Sprintf("Awaiting rename of original cluster")
	clusterInput := &rds.DescribeDBClustersInput{
		DBClusterIdentifier: aws_sdk.String(snapshotID),
	}
	retry.DoWithRetry(t, description, maxRetries, timeBetweenRetries, func() (string, error) {
		result, err := service.DescribeDBClusters(clusterInput)
		if err != nil {
			return "", err
		}
		json := fmt.Sprintf("%v", result)
		if !strings.Contains(json, `Status: "available"`) {
			return fmt.Sprintf("%v", result), fmt.Errorf("Still renaming")
		}
		return fmt.Sprintf("%v", result), nil
	})

	// Restore snapshot to a new cluster with the original identifier
	var securityGroupIDs []*string
	for _, sg := range dbSecurityGroups {
		securityGroupIDs = append(securityGroupIDs, &sg)
	}
	restoreConfig := &rds.RestoreDBClusterFromSnapshotInput{
		DBClusterIdentifier: aws_sdk.String(id),
		SnapshotIdentifier:  aws_sdk.String(snapshotID),
		Engine:              aws_sdk.String("aurora-postgresql"),
		DBSubnetGroupName:   aws_sdk.String(dbSubnetGroupName),
		VpcSecurityGroupIds: securityGroupIDs,
	}
	restoreResult, err := service.RestoreDBClusterFromSnapshot(restoreConfig)
	assert.NoError(t, err)
	fmt.Printf("[DEBUG] Snapshot Restored: %v\n", restoreResult)

	// Delete the RDS Cluster snapshot
	_, err = service.DeleteDBClusterSnapshot(
		&rds.DeleteDBClusterSnapshotInput{
			DBClusterSnapshotIdentifier: aws_sdk.String(snapshotID),
		},
	)
	assert.NoError(t, err)

	// Delete restored cluster
	deleteInput := &rds.DeleteDBClusterInput{
		DBClusterIdentifier: aws_sdk.String(id),
		SkipFinalSnapshot:   aws_sdk.Bool(true),
	}
	deleteResult, err := service.DeleteDBCluster(deleteInput)
	assert.NoError(t, err)
	fmt.Printf("[DEBUG] Cluster Delete Requested: %v\n", deleteResult)

	// Wait for deletion to complete
	description = fmt.Sprintf("Awaiting deletion of restored cluster")
	clusterInput = &rds.DescribeDBClustersInput{
		DBClusterIdentifier: aws_sdk.String(id),
	}
	retry.DoWithRetry(t, description, maxRetries, timeBetweenRetries, func() (string, error) {
		result, err := service.DescribeDBClusters(clusterInput)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				case rds.ErrCodeDBClusterNotFoundFault:
					return fmt.Sprintf("%v", result), nil
				default:
					return fmt.Sprintf("%v", result), err
				}
			} else {
				return fmt.Sprintf("%v", result), err
			}
		}
		return fmt.Sprintf("%v", result), fmt.Errorf("Still deleting")
	})

	// Rename original cluster
	modifyInput = &rds.ModifyDBClusterInput{
		ApplyImmediately:       aws_sdk.Bool(true),
		DBClusterIdentifier:    aws_sdk.String(snapshotID),
		NewDBClusterIdentifier: aws_sdk.String(id),
	}
	modifyResult, err = service.ModifyDBCluster(modifyInput)
	assert.NoError(t, err)
	fmt.Printf("[DEBUG] Cluster Rename Requested: %v\n", modifyResult)

	// Wait for rename to complete
	description = fmt.Sprintf("Awaiting rename of original cluster")
	clusterInput = &rds.DescribeDBClustersInput{
		DBClusterIdentifier: aws_sdk.String(snapshotID),
	}
	retry.DoWithRetry(t, description, maxRetries, timeBetweenRetries, func() (string, error) {
		result, err := service.DescribeDBClusters(clusterInput)
		if err != nil {
			return "", err
		}
		json := fmt.Sprintf("%v", result)
		if !strings.Contains(json, `Status: "available"`) {
			return fmt.Sprintf("%v", result), fmt.Errorf("Still renaming")
		}
		return fmt.Sprintf("%v", result), nil
	})

}
