package constants

import (
	"path/filepath"
)

// Global variables... yes I know that great, but hey this is automation tasks! 😃

// TargetCLIVersionArtifact is the path to the cli-version.json file.
var TargetCLIVersionArtifact = filepath.Join(ArtifactDirectory, "cli-version.json")
