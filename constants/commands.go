package constants

//  Constants for Actions
const (
	Read           = "read"
	Create         = "create"
	Update         = "update"
	Rollback       = "rollback"
	Edit           = "edit"
	Delete         = "delete"
	Describe       = "describe"
	Clear          = "clear"
	List           = "list"
	ChangePassword = "change-password"
	Search         = "search"
	BustCache      = "bustcache"
	AddMember      = "add-members"
	DeleteMember   = "delete-members"
	Restore        = "restore"
	Ping           = "ping"
)

// Nouns
const (
	NounSecret          = "secret"
	NounPermission      = "permission"
	NounPolicy          = "policy"
	NounPolicies        = "policies"
	NounAuth            = "auth"
	NounToken           = "token"
	NounUser            = "user"
	NounWhoAmI          = "whoami"
	EvaluateFlag        = "eval"
	Init                = "init"
	NounClient          = "client"
	NounAwsProfile      = "awsprofile"
	NounCliConfig       = "cli-config"
	NounConfig          = "config"
	NounRole            = "role"
	NounUsage           = "usage"
	NounGroup           = "group"
	NounAuthProvider    = "auth-provider"
	NounLogs            = "logs"
	NounAudit           = "audit"
	NounPrincipal       = "principal"
	NounPki             = "pki"
	NounSiem            = "siem"
	NounHome            = "home"
	NounPool            = "pool"
	NounEngine          = "engine"
	NounBootstrapUrl    = "url"
	NounBootstrapurlTTL = "url-ttl"
	NounClientUses      = "uses"
	NounClientTTL       = "ttl"
	NounClientDesc      = "desc"
)

// Cli-Config only
const (
	Editor = "editor"
)

// Flags
const (
	Key                     = "key"
	Value                   = "value"
	Profile                 = "profile"
	Path                    = "path"
	ID                      = "id"
	Data                    = "data"
	Username                = "auth.username"
	Tenant                  = "tenant"
	Password                = "auth.password"
	SecurePassword          = "auth.securePassword"
	CurrentPassword         = "currentPassword"
	NewPassword             = "newPassword"
	AuthProvider            = "auth.provider"
	Encoding                = "encoding"
	Beautify                = "beautify"
	Plain                   = "plain"
	Filter                  = "filter"
	Verbose                 = "verbose"
	Config                  = "config"
	Dev                     = "dev"
	AuthType                = "auth.type"
	AwsProfile              = "auth.awsprofile"
	GcpProject              = "auth.gcp.project"
	GcpToken                = "auth.gcp.token"
	GcpServiceAccount       = "auth.gcp.service"
	GcpAuthType             = "auth.gcp.type"
	AuthClientSecret        = "auth.client.secret"
	AuthClientID            = "auth.client.id"
	Callback                = "auth.callback"
	AzureAuthClientID       = "AZURE_CLIENT_ID"
	ThyOne                  = "thycoticone"
	ThyOneAuthClientBaseUri = "baseUri"
	ThyOneAuthClientID      = "clientId"
	ThyOneAuthClientSecret  = "clientSecret"
	SendWelcomeEmail        = "send.welcome.email"
	Query                   = "query"
	SearchLinks             = "search.links"
	SearchComparison        = "search.comparison"
	SearchType              = "search.type"
	SearchField             = "search.field"
	Limit                   = "limit"
	Cursor                  = "cursor"
	RefreshToken            = "refreshtoken"
	Output                  = "out"
	Overwrite               = "overwrite"
	ClientID                = "client.id"
	ClientSecret            = "client.secret"
	Version                 = "version"
	StartDate               = "startdate"
	EndDate                 = "enddate"
	Force                   = "force"
	Sort                    = "sort"
)

// Data Flags
const (
	// shared
	DataExternalID  = "external.id"
	DataDescription = "desc"
	DataAttributes  = "attributes"
	DataProvider    = "provider"

	// user
	DataUsername       = "username"
	DataSecurePassword = "securePassword"
	DataPassword       = "password"

	// permission
	DataSubject   = "subjects"
	DataEffect    = "effect"
	DataAction    = "actions"
	DataCondition = "conditions"
	DataCidr      = "cidr"
	DataResource  = "resources"

	// role and pool
	DataName     = "name"
	DataPoolName = "pool.name"

	// auth provider
	DataType      = "type"
	DataTenantID  = "azure.tenant.id"
	DataAccountID = "aws.account.id"
	DataProjectID = "gcp.project.id"
	DataCallback  = "callback"

	//group
	DataGroupName = "group.name"
	Members       = "members"
)

// Help Messages
const (
	CursorHelpMessage = "Next cursor for additional results. The cursor is provided at the end of each body response (\"cursor\": \"MQ==\") (optional)"
	ActionHelpMessage = "Action performed (POST, GET, PUT, PATCH or DELETE) (optional)"
)

// Security
const (
	StoreType = "store.type"
	Type      = "type"
	Store     = "store"
)

// Hidden Flags
const (
	StorePath = "store.path"
)

// Encodings
const (
	Yaml      = "yaml"
	Json      = "json"
	YamlShort = "yml"
)

func GetShortFlag(flag string) string {
	switch flag {
	case Tenant:
		return "t"
	case Config:
		return "c"
	default:
		return ""
	}
}
