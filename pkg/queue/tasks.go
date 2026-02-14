package queue

// Task Types
const (
	TaskTypeIndexPRD   = "index:prd"
	TaskTypeGenOutline = "gen:outline"
	TaskTypeGenCode    = "gen:code"
	TaskTypeGenDiagram = "gen:diagram"
)

// Payload for TaskTypeIndexPRD
type IndexPRDPayload struct {
	DocumentID string `json:"document_id"`
	SourceURL  string `json:"source_url"`
}

// Payload for TaskTypeGenCode
type GenCodePayload struct {
	TaskID      string `json:"task_id"`
	FeatureSpec string `json:"feature_spec"`
	Language    string `json:"language"`
}

// Payload for TaskTypeGenDiagram
type GenDiagramPayload struct {
	TaskID      string `json:"task_id"`
	Description string `json:"description"`
	DiagramType string `json:"diagram_type"` // e.g., "sequence", "class"
}
