package localfiles

const (
	BackupDirectory = "backups"
)

type LocalFilesRequirements struct {
	Path            string   `json:"file_path"`
	ExcludePatterns []string `json:"exclude_patterns,omitempty"`
}
