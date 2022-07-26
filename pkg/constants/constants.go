

package constants

const (
	APIVersion = "v1alpha1"

	KubeSystemNamespace           = "kube-system"
	OpenPitrixNamespace           = "openpitrix-system"
	CaptainDevOpsNamespace     = "captain-devops-system"
	IstioNamespace                = "istio-system"
	CaptainMonitoringNamespace = "Captain-monitoring-system"
	CaptainLoggingNamespace    = "Captain-logging-system"
	CaptainNamespace           = "Captain-system"
	CaptainControlNamespace    = "Captain-controls-system"
	PorterNamespace               = "porter-system"
	IngressControllerNamespace    = CaptainControlNamespace
	AdminUserName                 = "admin"
	IngressControllerPrefix       = "captain-router-"
	KubeSphereConfigName          = "captain-config"
	CaptainConfigMapDataKey       = "captain.yaml"

	ClusterNameLabelKey               = "captain.io/cluster"
	NameLabelKey                      = "captain.io/name"
	WorkspaceLabelKey                 = "captain.io/workspace"
	NamespaceLabelKey                 = "captain.io/namespace"
	DisplayNameAnnotationKey          = "captain.io/alias-name"
	ChartRepoIdLabelKey               = "captain.kubesphere.io/repo-id"
	ChartApplicationIdLabelKey        = "application.captain.io/app-id"
	ChartApplicationVersionIdLabelKey = "application.captain.io/app-version-id"
	CategoryIdLabelKey                = "application.captain.io/app-category-id"
	DanglingAppCleanupKey             = "application.captain.io/app-cleanup"
	CreatorAnnotationKey              = "captain.io/creator"
	UsernameLabelKey                  = "captain.io/username"
	DevOpsProjectLabelKey             = "captain.io/devopsproject"
	KubefedManagedLabel               = "kubefed.io/managed"

	UserNameHeader = "X-Token-Username"

	AuthenticationTag = "Authentication"
	UserTag           = "User"
	GroupTag          = "Group"

	WorkspaceMemberTag     = "Workspace Member"
	DevOpsProjectMemberTag = "DevOps Project Member"
	NamespaceMemberTag     = "Namespace Member"
	ClusterMemberTag       = "Cluster Member"

	GlobalRoleTag        = "Global Role"
	ClusterRoleTag       = "Cluster Role"
	WorkspaceRoleTag     = "Workspace Role"
	DevOpsProjectRoleTag = "DevOps Project Role"
	NamespaceRoleTag     = "Namespace Role"

	OpenpitrixTag            = "OpenPitrix Resources"
	OpenpitrixAppInstanceTag = "App Instance"
	OpenpitrixAppTemplateTag = "App Template"
	OpenpitrixCategoryTag    = "Category"
	OpenpitrixAttachmentTag  = "Attachment"
	OpenpitrixRepositoryTag  = "Repository"
	OpenpitrixManagementTag  = "App Management"
	// HelmRepoMinSyncPeriod min sync period in seconds
	HelmRepoMinSyncPeriod = 180

	CleanupDanglingAppOngoing = "ongoing"
	CleanupDanglingAppDone    = "done"

	DevOpsCredentialTag  = "DevOps Credential"
	DevOpsPipelineTag    = "DevOps Pipeline"
	DevOpsWebhookTag     = "DevOps Webhook"
	DevOpsJenkinsfileTag = "DevOps Jenkinsfile"
	DevOpsScmTag         = "DevOps Scm"
	DevOpsJenkinsTag     = "Jenkins"

	ToolboxTag      = "Toolbox"
	RegistryTag     = "Docker Registry"
	GitTag          = "Git"
	TerminalTag     = "Terminal"
	MultiClusterTag = "Multi-cluster"

	WorkspaceTag     = "Workspace"
	NamespaceTag     = "Namespace"
	DevOpsProjectTag = "DevOps Project"
	UserResourceTag  = "User's Resources"

	NamespaceResourcesTag = "Namespace Resources"
	ClusterResourcesTag   = "Cluster Resources"
	ComponentStatusTag    = "Component Status"

	GatewayTag = "Gateway"

	NetworkTopologyTag = "Network Topology"

	CaptainMetricsTag = "Captain Metrics"
	ClusterMetricsTag    = "Cluster Metrics"
	NodeMetricsTag       = "Node Metrics"
	NamespaceMetricsTag  = "Namespace Metrics"
	PodMetricsTag        = "Pod Metrics"
	PVCMetricsTag        = "PVC Metrics"
	IngressMetricsTag    = "Ingress Metrics"
	ContainerMetricsTag  = "Container Metrics"
	WorkloadMetricsTag   = "Workload Metrics"
	WorkspaceMetricsTag  = "Workspace Metrics"
	ComponentMetricsTag  = "Component Metrics"
	CustomMetricsTag     = "Custom Metrics"

	LogQueryTag      = "Log Query"
	EventsQueryTag   = "Events Query"
	AuditingQueryTag = "Auditing Query"

	ClusterMetersTag   = "Cluster Meters"
	NodeMetersTag      = "Node Meters"
	WorkspaceMetersTag = "Workspace Meters"
	NamespaceMetersTag = "Namespace Meters"
	WorkloadMetersTag  = "Workload Meters"
	PodMetersTag       = "Pod Meters"
	ServiceMetricsTag  = "ServiceName Meters"

	ApplicationReleaseName = "meta.helm.sh/release-name"
	ApplicationReleaseNS   = "meta.helm.sh/release-namespace"

	ApplicationName    = "app.kubernetes.io/name"
	ApplicationVersion = "app.kubernetes.io/version"
	AlertingTag        = "Alerting"

	NotificationTag             = "Notification"
	NotificationSecretNamespace = "captain-monitoring-federated"
	NotificationManagedLabel    = "notification.kubesphere.io/managed"

	DashboardTag = "Dashboard"
)

var (
	SystemNamespaces = []string{CaptainNamespace, CaptainLoggingNamespace, CaptainMonitoringNamespace, OpenPitrixNamespace, KubeSystemNamespace, IstioNamespace, CaptainDevOpsNamespace, PorterNamespace}
)
