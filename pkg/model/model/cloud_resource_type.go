package model

import "regexp"

type CloudResourceType string

func (c CloudResourceType) String() string {
	return string(c)
}

const (
	// AWS
	AWSAccount        CloudResourceType = "AWS::Organizations::Account"
	AWSLambdaFunction CloudResourceType = "AWS::Lambda::Function"
	AWSS3Bucket       CloudResourceType = "AWS::S3::Bucket"
	AWSEC2Instance    CloudResourceType = "AWS::EC2::Instance"
	AWSRole           CloudResourceType = "AWS::IAM::Role"
	AWSUser           CloudResourceType = "AWS::IAM::User"
	AWSGateway        CloudResourceType = "AWS::ApiGateway::RestApi"
	AWSSNSTopic       CloudResourceType = "AWS::SNS::Topic"
	AWSSQSQueue       CloudResourceType = "AWS::SQS::Queue"
	AWSRDSInstance    CloudResourceType = "AWS::RDS::DBInstance"

	// Azure
	AzureVM           CloudResourceType = "Microsoft.Compute/virtualMachines"
	AzureSubscription CloudResourceType = "Microsoft.Resources/subscriptions"

	// GCP
	GCPResourceBucket               CloudResourceType = "storage.googleapis.com/Bucket"
	GCPResourceInstance             CloudResourceType = "compute.googleapis.com/Instance"
	GCPResourceSQLInstance          CloudResourceType = "sqladmin.googleapis.com/Instance"
	GCPResourceFunction             CloudResourceType = "cloudfunctions.googleapis.com/Function"      // v1 and v2 functions
	GCPResourceFunctionV1           CloudResourceType = "cloudfunctions.googleapis.com/CloudFunction" // v1 functions - differences wrt reachability and triggers
	GCPResourceCloudRunJob          CloudResourceType = "run.googleapis.com/Job"
	GCPResourceCloudRunService      CloudResourceType = "run.googleapis.com/Service"
	GCPResourceAppEngineApplication CloudResourceType = "appengine.googleapis.com/Application"
	GCPResourceAppEngineService     CloudResourceType = "appengine.googleapis.com/Service"
	// GCP IAM Resources
	GCPResourceServiceAccount        CloudResourceType = "iam.googleapis.com/ServiceAccount"
	GCPResourceRole                  CloudResourceType = "iam.googleapis.com/Role"
	GCPResourcePolicy                CloudResourceType = "iam.googleapis.com/Policy"
	GCPResourceBinding               CloudResourceType = "iam.googleapis.com/Binding"
	GCPResourceMember                CloudResourceType = "iam.googleapis.com/Member"
	GCPResourceProject               CloudResourceType = "cloudresourcemanager.googleapis.com/Project"
	GCPResourceProjectPolicy         CloudResourceType = "cloudresourcemanager.googleapis.com/ProjectPolicy"
	GCPResourceProjectIamPolicy      CloudResourceType = "cloudresourcemanager.googleapis.com/ProjectIamPolicy"
	GCPResourceFolder                CloudResourceType = "cloudresourcemanager.googleapis.com/Folder"
	GCPResourceFolderPolicy          CloudResourceType = "cloudresourcemanager.googleapis.com/FolderPolicy"
	GCPResourceFolderIamPolicy       CloudResourceType = "cloudresourcemanager.googleapis.com/FolderIamPolicy"
	GCPResourceOrganization          CloudResourceType = "cloudresourcemanager.googleapis.com/Organization"
	GCPResourceOrganizationIamPolicy CloudResourceType = "cloudresourcemanager.googleapis.com/OrganizationIamPolicy"
	GCPResourceOrganizationPolicy    CloudResourceType = "cloudresourcemanager.googleapis.com/OrganizationPolicy"
	// Asset-only types (these produce assets, not likely to be used as resource types)
	GCPResourceForwardingRule       CloudResourceType = "compute.googleapis.com/ForwardingRule"
	GCPResourceGlobalForwardingRule CloudResourceType = "compute.googleapis.com/GlobalForwardingRule"
	GCPResourceDNSManagedZone       CloudResourceType = "dns.googleapis.com/ManagedZone"
	GCPResourceAddress              CloudResourceType = "compute.googleapis.com/Address" // used for both global and regional

	// Unknown - Catch all
	ResourceTypeUnknown CloudResourceType = "Unknown"
)

// Additional labels for each resource type
var resourceLabels = map[CloudResourceType][]string{
	AWSRole: {"Role", "Principal"},
	AWSUser: {"User", "Principal"},
}

var resourceValidators = map[CloudResourceType]*regexp.Regexp{
	AWSLambdaFunction: regexp.MustCompile(`^arn:aws:lambda:[a-z-0-9]+:\d{12}:function:.*$`),
	AWSS3Bucket:       regexp.MustCompile(`^arn:aws:s3:::.*$`),
	AWSEC2Instance:    regexp.MustCompile(`^arn:aws:ec2:[a-z-0-9]+:\d{12}:instance/.*$`),
	AWSRole:           regexp.MustCompile(`^arn:aws:iam::\d{12}:role/.*$`),
	AWSUser:           regexp.MustCompile(`^arn:aws:iam::\d{12}:user/.*$`),
	AWSGateway:        regexp.MustCompile(`^arn:aws:apigateway:[a-z-0-9]+:\d{12}:restapi:.*$`),
}
