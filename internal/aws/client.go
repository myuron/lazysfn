// Package aws provides AWS Step Functions API integration.
package aws

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sfn"
	"github.com/aws/aws-sdk-go-v2/service/sfn/types"
)

// SFNClient defines the subset of AWS SFN SDK methods used by this application.
type SFNClient interface {
	ListStateMachines(ctx context.Context, params *sfn.ListStateMachinesInput, optFns ...func(*sfn.Options)) (*sfn.ListStateMachinesOutput, error)
	ListExecutions(ctx context.Context, params *sfn.ListExecutionsInput, optFns ...func(*sfn.Options)) (*sfn.ListExecutionsOutput, error)
	GetExecutionHistory(ctx context.Context, params *sfn.GetExecutionHistoryInput, optFns ...func(*sfn.Options)) (*sfn.GetExecutionHistoryOutput, error)
	DescribeExecution(ctx context.Context, params *sfn.DescribeExecutionInput, optFns ...func(*sfn.Options)) (*sfn.DescribeExecutionOutput, error)
}

// Service provides business logic for fetching Step Functions data.
type Service struct {
	Client         SFNClient
	MaxConcurrency int
}

// StateMachine represents a Step Functions state machine with its latest execution status.
type StateMachine struct {
	Name         string
	ARN          string
	LatestStatus string
}

// Execution represents a single execution of a state machine.
type Execution struct {
	ID          string
	Status      string
	FailedState string
	StartTime   time.Time
	StopTime    time.Time
	InputParam  string
}

// FetchStateMachines retrieves all STANDARD type state machines, handling pagination.
func (s *Service) FetchStateMachines(ctx context.Context) ([]StateMachine, error) {
	var machines []StateMachine
	var nextToken *string

	for {
		input := &sfn.ListStateMachinesInput{
			NextToken: nextToken,
		}
		out, err := s.Client.ListStateMachines(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("listing state machines: %w", err)
		}
		for _, sm := range out.StateMachines {
			if sm.Type != types.StateMachineTypeStandard {
				continue
			}
			name := ""
			if sm.Name != nil {
				name = *sm.Name
			}
			arn := ""
			if sm.StateMachineArn != nil {
				arn = *sm.StateMachineArn
			}
			machines = append(machines, StateMachine{
				Name: name,
				ARN:  arn,
			})
		}
		if out.NextToken == nil {
			break
		}
		nextToken = out.NextToken
	}

	return machines, nil
}

// FetchLatestStatus enriches state machines with their latest execution status.
func (s *Service) FetchLatestStatus(ctx context.Context, machines []StateMachine) ([]StateMachine, error) {
	result := make([]StateMachine, len(machines))
	copy(result, machines)

	sem := make(chan struct{}, s.MaxConcurrency)
	var mu sync.Mutex
	var firstErr error
	var wg sync.WaitGroup

	for i := range result {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			out, err := s.Client.ListExecutions(ctx, &sfn.ListExecutionsInput{
				StateMachineArn: &result[idx].ARN,
				MaxResults:      1,
			})
			if err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = fmt.Errorf("listing executions for %s: %w", result[idx].Name, err)
				}
				mu.Unlock()
				return
			}
			if len(out.Executions) > 0 {
				result[idx].LatestStatus = string(out.Executions[0].Status)
			}
		}(i)
	}

	wg.Wait()
	if firstErr != nil {
		return nil, firstErr
	}
	return result, nil
}

// failureEventTypes is the set of event types that indicate a failure.
var failureEventTypes = map[types.HistoryEventType]bool{
	types.HistoryEventTypeTaskFailed:           true,
	types.HistoryEventTypeExecutionFailed:      true,
	types.HistoryEventTypeLambdaFunctionFailed: true,
	types.HistoryEventTypeActivityFailed:       true,
	types.HistoryEventTypeMapRunFailed:         true,
}

// FetchExecutionHistory retrieves recent executions for a state machine.
func (s *Service) FetchExecutionHistory(ctx context.Context, stateMachineARN string) ([]Execution, error) {
	listOut, err := s.Client.ListExecutions(ctx, &sfn.ListExecutionsInput{
		StateMachineArn: &stateMachineARN,
		MaxResults:      20,
	})
	if err != nil {
		return nil, fmt.Errorf("listing executions: %w", err)
	}

	executions := make([]Execution, len(listOut.Executions))
	sem := make(chan struct{}, s.MaxConcurrency)
	var mu sync.Mutex
	var firstErr error
	var wg sync.WaitGroup

	for i, item := range listOut.Executions {
		exec := &executions[i]
		exec.ID = derefStr(item.Name)
		exec.Status = string(item.Status)
		if item.StartDate != nil {
			exec.StartTime = *item.StartDate
		}
		if item.StopDate != nil {
			exec.StopTime = *item.StopDate
		}

		arn := derefStr(item.ExecutionArn)
		status := item.Status

		wg.Add(1)
		go func(exec *Execution, arn string, status types.ExecutionStatus) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			// Fetch failed state for failure statuses
			if needsFailedState(status) {
				failedState, err := s.fetchFailedState(ctx, arn)
				if err != nil {
					mu.Lock()
					if firstErr == nil {
						firstErr = err
					}
					mu.Unlock()
					return
				}
				exec.FailedState = failedState
			}

			// Fetch input param
			descOut, err := s.Client.DescribeExecution(ctx, &sfn.DescribeExecutionInput{
				ExecutionArn: &arn,
			})
			if err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = fmt.Errorf("describing execution %s: %w", arn, err)
				}
				mu.Unlock()
				return
			}
			if descOut.Input != nil {
				exec.InputParam = *descOut.Input
			}
		}(exec, arn, status)
	}

	wg.Wait()
	if firstErr != nil {
		return nil, firstErr
	}
	return executions, nil
}

func needsFailedState(status types.ExecutionStatus) bool {
	return status == types.ExecutionStatusFailed ||
		status == types.ExecutionStatusTimedOut ||
		status == types.ExecutionStatusAborted
}

func (s *Service) fetchFailedState(ctx context.Context, executionARN string) (string, error) {
	histOut, err := s.Client.GetExecutionHistory(ctx, &sfn.GetExecutionHistoryInput{
		ExecutionArn: &executionARN,
		ReverseOrder: true,
	})
	if err != nil {
		return "", fmt.Errorf("getting execution history for %s: %w", executionARN, err)
	}

	// Build event index by ID
	eventByID := make(map[int64]types.HistoryEvent, len(histOut.Events))
	for _, e := range histOut.Events {
		eventByID[e.Id] = e
	}

	// Find the failure event and trace back to the state name
	for _, e := range histOut.Events {
		if !failureEventTypes[e.Type] {
			continue
		}
		// Walk previousEventId chain to find TaskStateEntered
		prevID := e.PreviousEventId
		for prevID != 0 {
			prev, ok := eventByID[prevID]
			if !ok {
				break
			}
			if prev.StateEnteredEventDetails != nil && prev.StateEnteredEventDetails.Name != nil {
				return *prev.StateEnteredEventDetails.Name, nil
			}
			prevID = prev.PreviousEventId
		}
		break // Only check the first failure event
	}

	return "", nil
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
