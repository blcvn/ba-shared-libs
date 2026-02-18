package entities

import (
	"time"
)

type WorkflowMode string

const (
	ModeChat     WorkflowMode = "CHAT"
	ModePipeline WorkflowMode = "PIPELINE"
)

type StepType string

const (
	StepThought     StepType = "THOUGHT"
	StepAction      StepType = "ACTION"
	StepObservation StepType = "OBSERVATION"
	StepQuestion    StepType = "QUESTION"
	StepFinal       StepType = "FINAL"
)

type Step struct {
	ID        string
	Type      StepType
	Content   string
	Artifact  interface{}
	Timestamp time.Time
}

type Session struct {
	ID        string
	UserID    string
	Mode      WorkflowMode
	Status    string
	Context   map[string]interface{}
	History   []Step
	CreatedAt time.Time
	UpdatedAt time.Time
}

type AgentTask struct {
	ID           string
	SessionID    string
	Mode         WorkflowMode
	CurrentStep  int
	TotalSteps   int
	Input        string
	Intermediate map[string]interface{}
}
