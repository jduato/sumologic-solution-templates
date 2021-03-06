package test

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"

	http_helper "github.com/gruntwork-io/terratest/modules/http-helper"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
)

var props = loadPropertiesFile("../sumologic.auto.tfvars")

//SumoLogic Environment URL
var sumologicURL = getSumologicURL()

var customValidation = func(statusCode int, body string) bool { return statusCode == 200 }
var headers = map[string]string{"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte(getProperty("sumo_access_id")+":"+getProperty("sumo_access_key")))}

// Main function, define stages and run.
func TestTerraformSumoLogic(t *testing.T) {
	t.Parallel()

	workingDir := "../"

	// A unique ID we can use to namespace resources so we don't clash with anything already in the Sumo Logic
	uniqueID := random.UniqueId()
	collectorName := fmt.Sprintf("Atlassian_%s", uniqueID)

	// At the end of the test, un-deploy the solution using Terraform
	defer test_structure.RunTestStage(t, "cleanup", func() {
		destroyTerraform(t, workingDir)
	})

	// Deploy the solution using Terraform
	test_structure.RunTestStage(t, "deploy", func() {
		deployTerraform(t, workingDir, collectorName)
	})

	// Validate that the resources are created in Sumo Logic
	test_structure.RunTestStage(t, "validateSumoLogic", func() {
		validateSumoLogicResources(t, workingDir)
	})

	// Validate that the resources are created in Atlassian Systems
	test_structure.RunTestStage(t, "validateAtlassian", func() {
		validateAtlassianResources(t, workingDir)
	})

}

func loadPropertiesFile(file string) map[string]string {
	props, err := ReadPropertiesFile(file)
	if err != nil {
		fmt.Println("Error while reading properties file " + err.Error())
	}
	return props
}

func getProperty(property string) string {
	return props[property]
}

func getSumologicURL() string {
	u, _ := url.Parse(getProperty("sumo_api_endpoint"))
	return u.Scheme + "://" + u.Host
}

// Validate that the Sumo Logic Resources have been created
func validateSumoLogicResources(t *testing.T, workingDir string) {

	// Load the Terraform Options saved by the earlier deploy_terraform stage
	terraformOptions := test_structure.LoadTerraformOptions(t, workingDir)
	// Get folder where the Apps are installed
	folderName := strings.Replace(terraform.Output(t, terraformOptions, "folder_name"), " ", "%20", -1)
	// Run `terraform output` to get the value of an output variable
	collectorID := terraform.Output(t, terraformOptions, "atlassian_collector_id")
	// Validate if the collector is created successfully
	validateSumoLogicCollector(t, terraformOptions, collectorID)
	// Validate if the folder is created successfully
	validateSumoLogicFolder(t, terraformOptions)
	// Validate if the Jira Server Source is created successfully
	validateSumoLogicJiraServerSource(t, terraformOptions, collectorID)
	// Validate if the Jira Cloud Source is created successfully
	validateSumoLogicJiraCloudSource(t, terraformOptions, collectorID)
	// Validate if the Bitbucket Source is created successfully
	validateSumoLogicBitbucketSource(t, terraformOptions, collectorID)
	// Validate if the Opsgenie Source is created successfully
	validateSumoLogicOpsgenieSource(t, terraformOptions, collectorID)
	// Validate if the Sumologic Opsgenie Webhook is created successfully
	validateSumoLogicOpsgenieWebhook(t, terraformOptions)
	// Validate if the Jira Cloud Webhook is created successfully
	validateSumoLogicJiraCloudWebhook(t, terraformOptions)
	// Validate if the Jira Server Webhook is created successfully
	validateSumoLogicJiraServerWebhook(t, terraformOptions)
	// Validate if the Jira Service Desk Webhook is created successfully
	validateSumoLogicJiraServiceDeskWebhook(t, terraformOptions)
	// Validate if the Atlassian App is installed
	validateSumoLogicAtlassianAppInstallation(t, terraformOptions, folderName)
	// Validate if the Jira Server App is installed
	validateSumoLogicJiraServerAppInstallation(t, terraformOptions, folderName)
	// Validate if the Jira Cloud App is installed
	validateSumoLogicJiraCloudAppInstallation(t, terraformOptions, folderName)
	// Validate if the Bitbucket App is installed
	validateSumoLogicBitbucketAppInstallation(t, terraformOptions, folderName)
	// Validate if the Opsgenie App is installed
	validateSumoLogicOpsgenieAppInstallation(t, terraformOptions, folderName)
	// Validate if the Bitbucket Field is added successfully
	validateSumoLogicBitbucketField(t, terraformOptions, folderName)
}

func validateSumoLogicCollector(t *testing.T, terraformOptions *terraform.Options, collectorID string) {
	// Verify that we get back a 200 OK
	http_helper.HTTPDoWithCustomValidation(t, "GET", fmt.Sprintf("%s/api/v1/collectors/%s", sumologicURL, collectorID), nil, headers, customValidation, nil)
}

func validateSumoLogicFolder(t *testing.T, terraformOptions *terraform.Options) {
	// Run `terraform output` to get the value of an output variable
	folderID := terraform.Output(t, terraformOptions, "folder_id")

	// Verify that we get back a 200 OK
	http_helper.HTTPDoWithCustomValidation(t, "GET", fmt.Sprintf("%s/api/v2/content/folders/%s", sumologicURL, folderID), nil, headers, customValidation, nil)
}

func validateSumoLogicJiraServerSource(t *testing.T, terraformOptions *terraform.Options, collectorID string) {
	// Run `terraform output` to get the value of an output variable
	sourceID := terraform.Output(t, terraformOptions, "jira_server_source_id")
	if sourceID != "[]" && getProperty("install_jira_server") == "true" {
		sourceID = strings.Split(sourceID, "\"")[1]

		// Verify that we get back a 200 OK
		http_helper.HTTPDoWithCustomValidation(t, "GET", fmt.Sprintf("%s/api/v1/collectors/%s/sources/%s", sumologicURL, collectorID, sourceID), nil, headers, customValidation, nil)
	}
}

func validateSumoLogicJiraCloudSource(t *testing.T, terraformOptions *terraform.Options, collectorID string) {
	// Run `terraform output` to get the value of an output variable
	sourceID := terraform.Output(t, terraformOptions, "jira_cloud_source_id")
	if sourceID != "[]" && getProperty("install_jira_cloud") == "true" {
		sourceID = strings.Split(sourceID, "\"")[1]

		// Verify that we get back a 200 OK
		http_helper.HTTPDoWithCustomValidation(t, "GET", fmt.Sprintf("%s/api/v1/collectors/%s/sources/%s", sumologicURL, collectorID, sourceID), nil, headers, customValidation, nil)
	}
}

func validateSumoLogicBitbucketSource(t *testing.T, terraformOptions *terraform.Options, collectorID string) {
	// Run `terraform output` to get the value of an output variable
	sourceID := terraform.Output(t, terraformOptions, "bitbucket_cloud_source_id")
	if sourceID != "[]" && getProperty("install_bitbucket_cloud") == "true" {
		sourceID = strings.Split(sourceID, "\"")[1]

		// Verify that we get back a 200 OK
		http_helper.HTTPDoWithCustomValidation(t, "GET", fmt.Sprintf("%s/api/v1/collectors/%s/sources/%s", sumologicURL, collectorID, sourceID), nil, headers, customValidation, nil)
	}
}

func validateSumoLogicOpsgenieSource(t *testing.T, terraformOptions *terraform.Options, collectorID string) {
	// Run `terraform output` to get the value of an output variable
	sourceID := terraform.Output(t, terraformOptions, "opsgenie_source_id")
	if sourceID != "[]" && getProperty("install_opsgenie") == "true" {
		sourceID = strings.Split(sourceID, "\"")[1]

		// Verify that we get back a 200 OK
		http_helper.HTTPDoWithCustomValidation(t, "GET", fmt.Sprintf("%s/api/v1/collectors/%s/sources/%s", sumologicURL, collectorID, sourceID), nil, headers, customValidation, nil)
	}
}

func validateSumoLogicOpsgenieWebhook(t *testing.T, terraformOptions *terraform.Options) {
	// Run `terraform output` to get the value of an output variable
	webhookID := terraform.Output(t, terraformOptions, "sumo_opsgenie_webhook_id")
	if webhookID != "[]" && getProperty("install_sumo_to_opsgenie_webhook") == "true" {
		webhookID = strings.Split(webhookID, "\"")[1]
		// Verify that we get back a 200 OK
		http_helper.HTTPDoWithCustomValidation(t, "GET", fmt.Sprintf("%s/api/v1/connections/%s", sumologicURL, webhookID), nil, headers, customValidation, nil)
	}
}

func validateSumoLogicJiraCloudWebhook(t *testing.T, terraformOptions *terraform.Options) {
	// Run `terraform output` to get the value of an output variable
	webhookID := terraform.Output(t, terraformOptions, "sumo_jiracloud_webhook_id")
	if webhookID != "[]" && getProperty("install_sumo_to_jiracloud_webhook") == "true" {
		webhookID = strings.Split(webhookID, "\"")[1]
		// Verify that we get back a 200 OK
		http_helper.HTTPDoWithCustomValidation(t, "GET", fmt.Sprintf("%s/api/v1/connections/%s", sumologicURL, webhookID), nil, headers, customValidation, nil)
	}
}

func validateSumoLogicJiraServerWebhook(t *testing.T, terraformOptions *terraform.Options) {
	// Run `terraform output` to get the value of an output variable
	webhookID := terraform.Output(t, terraformOptions, "sumo_jiraserver_webhook_id")
	if webhookID != "[]" && getProperty("install_sumo_to_jiraserver_webhook") == "true" {
		webhookID = strings.Split(webhookID, "\"")[1]
		// Verify that we get back a 200 OK
		http_helper.HTTPDoWithCustomValidation(t, "GET", fmt.Sprintf("%s/api/v1/connections/%s", sumologicURL, webhookID), nil, headers, customValidation, nil)
	}
}

func validateSumoLogicJiraServiceDeskWebhook(t *testing.T, terraformOptions *terraform.Options) {
	// Run `terraform output` to get the value of an output variable
	webhookID := terraform.Output(t, terraformOptions, "sumo_jiraservicedesk_webhook_id")
	if webhookID != "[]" && getProperty("install_sumo_to_jiraservicedesk_webhook") == "true" {
		webhookID = strings.Split(webhookID, "\"")[1]
		// Verify that we get back a 200 OK
		http_helper.HTTPDoWithCustomValidation(t, "GET", fmt.Sprintf("%s/api/v1/connections/%s", sumologicURL, webhookID), nil, headers, customValidation, nil)
	}
}

func validateSumoLogicAtlassianAppInstallation(t *testing.T, terraformOptions *terraform.Options, folderName string) {
	if getProperty("install_atlassian_app") == "true" {
		appFolderPath := fmt.Sprintf("/Library/Users/%s/%s/Atlassian", strings.Replace(os.Getenv("SUMOLOGIC_USERNAME"), "+", "%2B", -1), folderName)
		// Verify that we get back a 200 OK
		http_helper.HTTPDoWithCustomValidation(t, "GET", fmt.Sprintf("%s/api/v2/content/path?path=%s", sumologicURL, appFolderPath), nil, headers, customValidation, nil)
	}
}

func validateSumoLogicJiraCloudAppInstallation(t *testing.T, terraformOptions *terraform.Options, folderName string) {

	if getProperty("install_jira_cloud") == "true" {
		appFolderPath := fmt.Sprintf("/Library/Users/%s/%s/Jira%%20Cloud", strings.Replace(os.Getenv("SUMOLOGIC_USERNAME"), "+", "%2B", -1), folderName)
		// Verify that we get back a 200 OK
		http_helper.HTTPDoWithCustomValidation(t, "GET", fmt.Sprintf("%s/api/v2/content/path?path=%s", sumologicURL, appFolderPath), nil, headers, customValidation, nil)
	}
}

func validateSumoLogicJiraServerAppInstallation(t *testing.T, terraformOptions *terraform.Options, folderName string) {

	if getProperty("install_jira_server") == "true" {
		appFolderPath := fmt.Sprintf("/Library/Users/%s/%s/Jira", strings.Replace(os.Getenv("SUMOLOGIC_USERNAME"), "+", "%2B", -1), folderName)
		// Verify that we get back a 200 OK
		http_helper.HTTPDoWithCustomValidation(t, "GET", fmt.Sprintf("%s/api/v2/content/path?path=%s", sumologicURL, appFolderPath), nil, headers, customValidation, nil)
	}
}

func validateSumoLogicBitbucketAppInstallation(t *testing.T, terraformOptions *terraform.Options, folderName string) {

	if getProperty("install_bitbucket_cloud") == "true" {
		appFolderPath := fmt.Sprintf("/Library/Users/%s/%s/Bitbucket", strings.Replace(os.Getenv("SUMOLOGIC_USERNAME"), "+", "%2B", -1), folderName)
		// Verify that we get back a 200 OK
		http_helper.HTTPDoWithCustomValidation(t, "GET", fmt.Sprintf("%s/api/v2/content/path?path=%s", sumologicURL, appFolderPath), nil, headers, customValidation, nil)
	}
}

func validateSumoLogicOpsgenieAppInstallation(t *testing.T, terraformOptions *terraform.Options, folderName string) {

	if getProperty("install_opsgenie") == "true" {
		appFolderPath := fmt.Sprintf("/Library/Users/%s/%s/Opsgenie", strings.Replace(os.Getenv("SUMOLOGIC_USERNAME"), "+", "%2B", -1), folderName)
		// Verify that we get back a 200 OK
		http_helper.HTTPDoWithCustomValidation(t, "GET", fmt.Sprintf("%s/api/v2/content/path?path=%s", sumologicURL, appFolderPath), nil, headers, customValidation, nil)
	}
}

func validateSumoLogicBitbucketField(t *testing.T, terraformOptions *terraform.Options, folderName string) {

	// Run `terraform output` to get the value of an output variable
	fieldID := terraform.Output(t, terraformOptions, "sumo_bitbucket_field_id")
	if fieldID != "[]" && getProperty("install_bitbucket_cloud") == "true" {
		fieldID = strings.Split(fieldID, "\"")[1]
		// Verify that we get back a 200 OK
		http_helper.HTTPDoWithCustomValidation(t, "GET", fmt.Sprintf("%s/api/v1/fields/%s", sumologicURL, fieldID), nil, headers, customValidation, nil)
	}
}

// Validate that the Atlassian Resources has been deployed
func validateAtlassianResources(t *testing.T, workingDir string) {

	// Load the Terraform Options saved by the earlier deploy_terraform stage
	terraformOptions := test_structure.LoadTerraformOptions(t, workingDir)
	validateAtlassianOpsgenieWebhook(t, terraformOptions)
	//validateAtlassianBitbucketWebhook(t, terraformOptions) - See method comment  as for why this  is commented.
	//validateAtlassianJiraServerWebhook(t, terraformOptions) - Commented because Webhook api's can be called only from connect apps. https://developer.atlassian.com/cloud/jira/platform/rest/v2/#api-rest-api-2-webhook-get
	//validateAtlassianJiraCloudWebhook(t, terraformOptions) - Commented because Webhook api's can be called only from connect apps. https://developer.atlassian.com/cloud/jira/platform/rest/v2/#api-rest-api-2-webhook-get
}

// Below function is commented because right now there is no way to identify workspace, repo and id mapping as the terraform creation returns  only id.
// https://github.com/terraform-providers/terraform-provider-bitbucket/blob/a677ca55116512a845b66e1d3df5973492d12328/bitbucket/resource_hook.go#L92
// https://developer.atlassian.com/bitbucket/api/2/reference/resource/repositories/%7Bworkspace%7D/%7Brepo_slug%7D
// func validateAtlassianBitbucketWebhook(t *testing.T, terraformOptions *terraform.Options) {
//  var bitbucketURL = "https://api.bitbucket.org"
//  var bitbucketheaders = map[string]string{"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte("bitbucket_username:bitbucket_password"))}
// 	// Run `terraform output` to get the value of an output variable
// 	webhookID := terraform.Output(t, terraformOptions, "bitbucket_webhook_id")
// 	if webhookID != "[]" {
// 		webhookID = strings.Replace(strings.Replace(strings.Replace(strings.Replace(strings.Replace(webhookID, "{", "", -1), "}", "", -1), "[", "", -1), "]", "", -1), "\"", "", -1)
// 		webhookIDs := strings.Split(webhookID, ",")
// 		for i := 0; i < len(webhookIDs); i++ {
// 			// Verify that we get back a 200 OK
// 			http_helper.HTTPDoWithCustomValidation(t, "GET", fmt.Sprintf("%s/2.0/repositories/<workspace>/<repo>/hooks/%s", bitbucketURL, strings.TrimSpace(webhookIDs[i])), nil, bitbucketheaders, customValidation, nil)
// 		}
// 	}
// }

func validateAtlassianOpsgenieWebhook(t *testing.T, terraformOptions *terraform.Options) {
	var opsProps = loadPropertiesFile("../atlassian.auto.tfvars")
	var opsgenieURL = opsProps["opsgenie_api_url"]
	var opsgenieheaders = map[string]string{"Authorization": "GenieKey " + opsProps["opsgenie_key"], "Content-Type": "application/json"}
	// Run `terraform output` to get the value of an output variable
	webhookID := terraform.Output(t, terraformOptions, "opsgenie_webhook_id")
	if webhookID != "[]" && getProperty("install_opsgenie") == "true" {
		webhookID = strings.Split(webhookID, "\"")[1]
		// Verify that we get back a 200 OK
		http_helper.HTTPDoWithCustomValidation(t, "GET", fmt.Sprintf("%s/v2/integrations/%s", opsgenieURL, webhookID), nil, opsgenieheaders, customValidation, nil)
	}
}

// Deploy the resources using Terraform
func deployTerraform(t *testing.T, workingDir string, collectorName string) {

	terraformOptions := &terraform.Options{
		// The path to where our Terraform code is located
		TerraformDir: workingDir,

		// Variables to pass to our Terraform code using -var options
		Vars: map[string]interface{}{
			"collector_name": collectorName,
		},

		// Disable colors in Terraform commands so its easier to parse stdout/stderr
		NoColor: true,
	}

	// Save the Terraform Options struct, instance name, and instance text so future test stages can use it
	test_structure.SaveTerraformOptions(t, workingDir, terraformOptions)

	// This will run `terraform init` and `terraform apply` and fail the test if there are any errors
	terraform.InitAndApply(t, terraformOptions)
}

// Destroy the resources using Terraform
func destroyTerraform(t *testing.T, workingDir string) {
	// Load the Terraform Options saved by the earlier deploy_terraform stage
	terraformOptions := test_structure.LoadTerraformOptions(t, workingDir)

	terraform.Destroy(t, terraformOptions)
}
