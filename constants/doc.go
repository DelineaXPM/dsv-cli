package constants

// Constants used for ducomentation
const (
	ExamplePath             = "databases/mongo-db01"
	ExamplePolicyPath       = "secrets/databases/mongo-db01"
	ExampleRoleName         = "gcp-svc-1"
	ExampleAuthProviderName = "aws-dev"
	ExamplePath2            = "cloudservices/aws/ec2-provision"
	ExampleDataJSON         = `'{"Key":"Value","Key2":"Value2"}'`
	ExampleDataPath         = "@/tmp/data.json"
	ExampleConfigPath       = "@/tmp/config.yml"
	ExampleUser             = "kadmin"
	ExampleSIEM             = "LogsInc"
	ExamplePassword         = "********"
	ExampleUserSearch       = "adm"
	ExamplePolicySearch     = "secrets/databases"
	ExampleAuthClientID     = "8sdh2el7-ai29S05a5"
	ExampleAuthClientSecret = "pLaKNWL99IK2kL-xMI"
	ExampleClientID         = "8sdh2el7-ai29S05a5"
	ExampleClientSecret     = "pLaKNWL99IK2kL-xMI"
	ExampleGroup            = "administrators"
	ExampleGroupCreate      = `{"groupName": "administrators"}`
	ExampleGroupAddMembers  = `{"members": ["member1","member2"]}`
	GCPNote                 = `GCP GCE metadata auth provider can be created in the command line, but a GCP Service Account must be done using a file.  See the Authentication:GCP portion of the documentation.`
)
