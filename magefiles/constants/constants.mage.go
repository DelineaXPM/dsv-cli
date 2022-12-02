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
