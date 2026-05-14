package domain

type VersionKey struct {
	ResourceID      uint
	ResourceVersion string
	PromptVersion   string
	SchemaVersion   string
	ModelVersion    string
}

func (k VersionKey) Equal(other VersionKey) bool {
	return k.ResourceID == other.ResourceID &&
		k.ResourceVersion == other.ResourceVersion &&
		k.PromptVersion == other.PromptVersion &&
		k.SchemaVersion == other.SchemaVersion &&
		k.ModelVersion == other.ModelVersion
}

func (k VersionKey) Complete() bool {
	return k.ResourceID != 0 &&
		k.ResourceVersion != "" &&
		k.PromptVersion != "" &&
		k.SchemaVersion != "" &&
		k.ModelVersion != ""
}
