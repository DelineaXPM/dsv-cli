package constants

const (
	ProductName = "DevOps Secrets Vault"
)

// Constants used for documentation.
const (
	ExamplePath             = "databases/mongo-db01"
	ExamplePolicyPath       = "secrets/databases/mongo-db01"
	ExampleRoleName         = "gcp-svc-1"
	ExampleDataJSON         = `'{"Key":"Value","Key2":"Value2"}'`
	ExampleDataPath         = "@/tmp/data.json"
	ExampleBodyDataJSON     = `'{"data":{"Key":"Value"},"attributes":{"Key2":"Value2"},"desc":"Secret description"}'`
	ExampleConfigPath       = "@/tmp/config.yml"
	ExampleUser             = "kadmin"
	ExampleSIEM             = "LogsInc"
	ExampleAuthType         = "clientcred"
	ExampleUserSearch       = "adm"
	ExamplePolicySearch     = "secrets/databases"
	ExampleSiemSearch       = "my_siem"
	ExampleAuthClientID     = "8sdh2el7-ai29S05a5"
	ExampleAuthClientSecret = "pLaKNWL99IK2kL-xMI"
	ExampleDomain           = "secretsvaultcloud.com"
	ExampleClientID         = "8sdh2el7-ai29S05a5"
	ExampleGroup            = "administrators"
	ExampleGroupCreate      = `{"groupName": "administrators"}`
	ExampleGroupAddMembers  = `{"members": ["member1","member2"]}`
)
