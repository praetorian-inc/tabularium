package model

import (
	"regexp"
	"strings"
)

type CloudResourceType string

func (c CloudResourceType) String() string {
	return string(c)
}

const (
	// AWS
	AWSAccount             CloudResourceType = "AWS::Organizations::Account"
	AWSOrganization        CloudResourceType = "AWS::Organizations::Organization"
	AWSLambdaFunction      CloudResourceType = "AWS::Lambda::Function"
	AWSS3Bucket            CloudResourceType = "AWS::S3::Bucket"
	AWSEC2Instance         CloudResourceType = "AWS::EC2::Instance"
	AWSCloudFormationStack CloudResourceType = "AWS::CloudFormation::Stack"
	AWSEcrRepository       CloudResourceType = "AWS::ECR::Repository"
	AWSEcrPublicRepository CloudResourceType = "AWS::ECR::PublicRepository"
	AWSRole                CloudResourceType = "AWS::IAM::Role"
	AWSUser                CloudResourceType = "AWS::IAM::User"
	AWSGroup               CloudResourceType = "AWS::IAM::Group"
	AWSManagedPolicy       CloudResourceType = "AWS::IAM::ManagedPolicy"
	AWSGateway             CloudResourceType = "AWS::ApiGateway::RestApi"
	AWSSNSTopic            CloudResourceType = "AWS::SNS::Topic"
	AWSSQSQueue            CloudResourceType = "AWS::SQS::Queue"
	AWSRDSInstance         CloudResourceType = "AWS::RDS::DBInstance"
	AWSServicePrincipal    CloudResourceType = "AWS::IAM::ServicePrincipal"
	AWSEC2SecurityGroup    CloudResourceType = "AWS::EC2::SecurityGroup"
	AWSEC2NetworkAcl       CloudResourceType = "AWS::EC2::NetworkAcl"
	AWSEC2Volume           CloudResourceType = "AWS::EC2::Volume"
	AWSEC2VPCEndpoint      CloudResourceType = "AWS::EC2::VPCEndpoint"
	AWSEC2LaunchTemplate   CloudResourceType = "AWS::EC2::LaunchTemplate"
	AWSEC2Subnet           CloudResourceType = "AWS::EC2::Subnet"
	AWSVPC                 CloudResourceType = "AWS::EC2::VPC"
	AWSELB                 CloudResourceType = "AWS::ElasticLoadBalancingV2::LoadBalancer"
	AWSEKSCluster          CloudResourceType = "AWS::EKS::Cluster"
	AWSRDSSnapshot         CloudResourceType = "AWS::RDS::DBSnapshot"
	AWSRDSCluster          CloudResourceType = "AWS::RDS::DBCluster"
	AWSRDSClusterSnapshot  CloudResourceType = "AWS::RDS::DBClusterSnapshot"
	AWSElastiCacheCluster  CloudResourceType = "AWS::ElastiCache::ReplicationGroup"
	AWSCloudTrail          CloudResourceType = "AWS::CloudTrail::Trail"
	AWSCloudFront          CloudResourceType = "AWS::CloudFront::Distribution"
	AWSDynamoDBTable       CloudResourceType = "AWS::DynamoDB::Table"
	AWSGlueCatalog         CloudResourceType = "AWS::Glue::DataCatalog"
	AWSKinesisStream       CloudResourceType = "AWS::Kinesis::Stream"
	AWSECSTaskDefinition   CloudResourceType = "AWS::ECS::TaskDefinition"

	// Azure
	AzureVM                       CloudResourceType = "Microsoft.Compute/virtualMachines"
	AzureSubscription             CloudResourceType = "Microsoft.Resources/subscriptions"
	AzureVMUserData               CloudResourceType = "Microsoft.Compute/virtualMachines/userData"
	AzureVMExtensions             CloudResourceType = "Microsoft.Compute/virtualMachines/extensions"
	AzureVMDiskEncryption         CloudResourceType = "Microsoft.Compute/virtualMachines/diskEncryption"
	AzureVMTags                   CloudResourceType = "Microsoft.Compute/virtualMachines/tags"
	AzureWebSite                  CloudResourceType = "Microsoft.Web/Sites"
	AzureWebSiteConfiguration     CloudResourceType = "Microsoft.Web/sites/configuration"
	AzureWebSiteConnectionStrings CloudResourceType = "Microsoft.Web/sites/connectionStrings"
	AzureWebSiteKeys              CloudResourceType = "Microsoft.Web/sites/keys"
	AzureWebSiteSettings          CloudResourceType = "Microsoft.Web/sites/settings"
	AzureWebSiteTags              CloudResourceType = "Microsoft.Web/sites/tags"
	AzureAutomationRunbooks       CloudResourceType = "Microsoft.Automation/automationAccounts/runbooks"
	AzureAutomationVariables      CloudResourceType = "Microsoft.Automation/automationAccounts/variables"
	AzureAutomationJobs           CloudResourceType = "Microsoft.Automation/automationAccounts/jobs"

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
	GCRContainerImage               CloudResourceType = "containerregistry.googleapis.com/Image"
	GCRArtifactRepository           CloudResourceType = "artifactregistry.googleapis.com/Repository"
	GCRArtifactoryDockerImage       CloudResourceType = "artifactregistry.googleapis.com/DockerImage"

	// Kubernetes
	K8sDeployment     CloudResourceType = "k8s.io/Deployment"
	K8sStatefulSet    CloudResourceType = "k8s.io/StatefulSet"
	K8sDaemonSet      CloudResourceType = "k8s.io/DaemonSet"
	K8sJob            CloudResourceType = "k8s.io/Job"
	K8sCronJob        CloudResourceType = "k8s.io/CronJob"
	K8sPod            CloudResourceType = "k8s.io/Pod"
	K8sService        CloudResourceType = "k8s.io/Service"
	K8sSecret         CloudResourceType = "k8s.io/Secret"
	K8sServiceAccount CloudResourceType = "k8s.io/ServiceAccount"
	K8sClusterRole    CloudResourceType = "k8s.io/ClusterRole"
	K8sRollout        CloudResourceType = "k8s.io/Rollout"
	K8sIngress        CloudResourceType = "k8s.io/Ingress"

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

func IsSupportedResourceType(tipe CloudResourceType, supported []CloudResourceType) bool {
	for _, s := range supported {
		if strings.EqualFold(tipe.String(), s.String()) {
			return true
		}

	}

	return false
}
