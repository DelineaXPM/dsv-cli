package constants

// Constants for Actions
const (
	Read           = "read"
	Create         = "create"
	Update         = "update"
	Rollback       = "rollback"
	Upload         = "upload"
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
	Rotate         = "rotate"
	Encrypt        = "encrypt"
	Decrypt        = "decrypt"
	Generate       = "generate"
	Apply          = "apply"
	Status         = "status"
	UseProfile     = "use-profile"
)

// Nouns
const (
	NounSecret          = "secret"
	NounSecrets         = "secrets"
	NounPolicy          = "policy"
	NounPolicies        = "policies"
	NounAuth            = "auth"
	NounToken           = "token"
	NounUser            = "user"
	NounUsers           = "users"
	NounWhoAmI          = "whoami"
	EvaluateFlag        = "eval"
	Init                = "init"
	NounClient          = "client"
	NounClients         = "clients"
	NounAwsProfile      = "awsprofile"
	NounCliConfig       = "cli-config"
	NounConfig          = "config"
	NounRole            = "role"
	NounRoles           = "roles"
	NounUsage           = "usage"
	NounGroup           = "group"
	NounGroups          = "groups"
	NounAuthProvider    = "auth-provider"
	NounLogs            = "logs"
	NounAudit           = "audit"
	NounPrincipal       = "principal"
	NounPki             = "pki"
	NounSiem            = "siem"
	NounSiems           = "siems"
	NounHome            = "home"
	NounPool            = "pool"
	NounPools           = "pools"
	NounEngine          = "engine"
	NounEngines         = "engines"
	NounBootstrapUrl    = "url"
	NounBootstrapUrlTTL = "url-ttl"
	NounClientUses      = "uses"
	NounClientTTL       = "ttl"
	NounClientDesc      = "desc"
	NounKey             = "key"
	NounEncryption      = "crypto"
	NounReport          = "report"
	NounBreakGlass      = "breakglass"
	NounCert            = "certificate"
	NounPrivateKey      = "privateKey"
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
	DomainName              = "domain"
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
	AuthCert                = "auth.certificate"
	AuthPrivateKey          = "auth.privateKey"
	ThyOne                  = "thycoticone"
	ThyOneAuthClientBaseUri = "baseUri"
	ThyOneAuthClientID      = "clientId"
	ThyOneAuthClientSecret  = "clientSecret"
	SendWelcomeEmail        = "send.welcome.email"

	Query             = "query"
	SearchLinks       = "search.links"
	SearchComparison  = "search.comparison"
	SearchType        = "search.type"
	SearchField       = "search.field"
	Limit             = "limit"
	OffSet            = "offset"
	Cursor            = "cursor"
	RefreshToken      = "refreshtoken"
	Output            = "out"
	Overwrite         = "overwrite"
	ClientID          = "client.id"
	ClientSecret      = "client.secret"
	Version           = "version"
	VersionStart      = "version-start"
	VersionEnd        = "version-end"
	StartDate         = "startdate"
	EndDate           = "enddate"
	Force             = "force"
	Sort              = "sort"
	NewAdmins         = "new-admins"
	MinNumberOfShares = "min-number-of-shares"
	Shares            = "shares"
	SendToEngine      = "send-to-engine"
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
	DataDisplayname    = "displayname"

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

// Common Help Messages
const (
	CursorHelpMessage = "Next cursor for additional results. The cursor is provided at the end of each body response (\"cursor\": \"MQ==\") (optional)"
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

// Control authentication cache usage.
const (
	AuthSkipCache = "auth.skip.cache"
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
