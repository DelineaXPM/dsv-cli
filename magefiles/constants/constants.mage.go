package constants

// Since we are dealing with builds, having a constants file until using a config input makes it easy.

const (
	// ArtifactDirectory is a directory containing artifacts for the project and shouldn't be committed to source.
	ArtifactDirectory = ".artifacts"

	// PermissionUserReadWriteExecute is the permissions for the artifact directory.
	PermissionUserReadWriteExecute = 0o0700

	// CacheDirectory is where the cache for the project is placed, ie artifacts that don't need to be rebuilt often.
	CacheDirectory = ".cache"
)

// Publishing constants
const (
	// S3CLIVersionPath is the S3 fully qualified key location to upload the cli-versions.json file.
	S3CLIVersionPath = "cli-version.json"

	// AWSDefaultS3Region is the default region for S3 buckets, at this time a single configured value for us-east-1.
	AWSDefaultS3Region = "us-east-1"
)
