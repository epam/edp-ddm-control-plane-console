package query

import "time"

const (
	viewTimeFormat = "02.01.2006 15:04"
	statusActive   = "active"
)

type Codebase struct {
	ID                   int               `json:"id" orm:"column(id)"`
	Name                 string            `json:"name" orm:"column(name)"`
	Language             string            `json:"language" orm:"column(language)"`
	BuildTool            string            `json:"build_tool" orm:"column(build_tool)"`
	Framework            string            `json:"framework" orm:"column(framework)"`
	Strategy             string            `json:"strategy" orm:"column(strategy)"`
	GitURL               string            `json:"git_url" orm:"column(repository_url)"`
	RouteSite            string            `json:"route_site" orm:"column(route_site)"`
	RoutePath            string            `json:"route_path" orm:"column(route_path)"`
	DbKind               string            `json:"db_kind" orm:"column(database_kind)"`
	DbVersion            string            `json:"db_version" orm:"column(database_version)"`
	DbCapacity           string            `json:"db_capacity" orm:"column(database_capacity)"`
	DbStorage            string            `json:"db_storage" orm:"column(database_storage)"`
	Type                 CodebaseType      `json:"type" orm:"column(type)"`
	Status               Status            `json:"status" orm:"column(status)"`
	TestReportFramework  string            `json:"testReportFramework" orm:"column(test_report_framework)"`
	Description          string            `json:"description" orm:"column(description)"`
	CodebaseBranch       []*CodebaseBranch `json:"codebase_branch" orm:"reverse(many)"`
	ActionLog            []*ActionLog      `json:"-" orm:"rel(m2m);rel_table(codebase_action_log)"`
	GitServerID          *int              `json:"-" orm:"column(git_server_id)"`
	GitServer            *string           `json:"gitServer" orm:"-"`
	GitProjectPath       *string           `json:"gitProjectPath" orm:"column(git_project_path)"`
	JenkinsSlaveID       *int              `json:"-" orm:"column(jenkins_slave_id)"`
	JenkinsSlave         string            `json:"jenkinsSlave" orm:"-"`
	JobProvisioningID    *int              `json:"-" orm:"column(job_provisioning_id)"`
	JobProvisioning      string            `json:"jobProvisioning" orm:"-"`
	DeploymentScript     string            `json:"deploymentScript" orm:"deployment_script"`
	VersioningType       string            `json:"versioningType" orm:"versioning_type"`
	StartVersioningFrom  *string           `json:"startFrom" orm:"start_versioning_from"`
	JiraServerID         *int              `json:"-" orm:"column(jira_server_id)"`
	JiraServer           *string           `json:"jiraServer" orm:"-"`
	CommitMessagePattern string            `json:"commitMessagePattern" orm:"commit_message_pattern"`
	TicketNamePattern    string            `json:"ticketNamePattern" orm:"ticket_name_pattern"`
	CiTool               string            `json:"ciTool" orm:"ci_tool"`
	CreatedAt            *time.Time        `json:"-" orm:"-"`
	ForegroundDeletion   bool              `json:"-" orm:"-"`
	Available            bool              `json:"-" orm:"-"`
	Admins               string            `json:"-"`
}

func (c Codebase) FormattedCreatedAt() string {
	if c.CreatedAt != nil {
		return c.CreatedAt.Format(viewTimeFormat)
	}

	if len(c.ActionLog) == 0 {
		return ""
	}

	return c.ActionLog[0].LastTimeUpdate.Format(viewTimeFormat)
}

func (c Codebase) StrStatus() string {
	status := string(c.Status)
	if status == "" {
		status = "active"
	}

	return status
}

func (c Codebase) CanBeDeleted() bool {
	for _, cb := range c.CodebaseBranch {
		if cb.Status != statusActive {
			return false
		}
	}

	return c.Available && string(c.Status) == statusActive
}

func (c *Codebase) TableName() string {
	return "codebase"
}

type CodebaseCriteria struct {
	BranchStatus Status
	Status       Status
	Type         CodebaseType
	Language     CodebaseLanguage
}

type CodebaseType string
type CodebaseLanguage string

const (
	App       CodebaseType = "application"
	Autotests CodebaseType = "autotests"
	Library   CodebaseType = "library"
	Registry  CodebaseType = "library" // temporary needs, change to registry-tenant
)

var CodebaseTypes = map[string]CodebaseType{
	"application": App,
	"autotests":   Autotests,
	"library":     Library,
}

type ApplicationsToPromote struct {
	ID           int `orm:"column(id)"`
	CdPipelineID int `orm:"column(cd_pipeline_id)"`
	CodebaseID   int `orm:"column(codebase_id)"`
}

func (c *ApplicationsToPromote) TableName() string {
	return "applications_to_promote"
}
