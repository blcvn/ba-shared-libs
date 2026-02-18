package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// TasksTotal tracks total tasks executed
	TasksTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ba_agent_tasks_total",
			Help: "Total number of tasks executed",
		},
		[]string{"status"}, // status: completed, failed, cancelled
	)

	// TaskDuration tracks task execution time
	TaskDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ba_agent_task_duration_seconds",
			Help:    "Task execution duration in seconds",
			Buckets: []float64{1, 5, 10, 30, 60, 120, 300},
		},
		[]string{"status"},
	)

	// ReActIterations tracks number of ReAct loop iterations
	ReActIterations = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "ba_agent_react_iterations",
			Help:    "Number of ReAct loop iterations per task",
			Buckets: []float64{1, 3, 5, 10, 15, 20, 30},
		},
	)

	// ToolCalls tracks tool execution count
	ToolCalls = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ba_agent_tool_calls_total",
			Help: "Total number of tool calls",
		},
		[]string{"tool_name", "status"}, // status: success, error
	)

	// LLMCalls tracks LLM API calls
	LLMCalls = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ba_agent_llm_calls_total",
			Help: "Total number of LLM API calls",
		},
		[]string{"model_id", "status"},
	)

	// TokensUsed tracks token consumption
	TokensUsed = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "ba_agent_tokens_used_total",
			Help: "Total tokens used",
		},
	)

	// CostTotal tracks cumulative cost
	CostTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "ba_agent_cost_total",
			Help: "Total cost in USD",
		},
	)
)
