package aws

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
	"github.com/aws/aws-sdk-go-v2/service/sfn/types"
)

// --- FetchStateMachines tests ---

func TestFetchStateMachines_SinglePage(t *testing.T) {
	mock := &mockSFNClient{
		listStateMachinesFn: func(ctx context.Context, params *sfn.ListStateMachinesInput, optFns ...func(*sfn.Options)) (*sfn.ListStateMachinesOutput, error) {
			return &sfn.ListStateMachinesOutput{
				StateMachines: []types.StateMachineListItem{
					{Name: aws.String("sm-1"), StateMachineArn: aws.String("arn:aws:states:us-east-1:123:stateMachine:sm-1"), Type: types.StateMachineTypeStandard},
					{Name: aws.String("sm-2"), StateMachineArn: aws.String("arn:aws:states:us-east-1:123:stateMachine:sm-2"), Type: types.StateMachineTypeStandard},
				},
			}, nil
		},
	}
	svc := &Service{Client: mock}
	machines, err := svc.FetchStateMachines(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(machines) != 2 {
		t.Fatalf("expected 2 machines, got %d", len(machines))
	}
	if machines[0].Name != "sm-1" {
		t.Errorf("machines[0].Name = %q, want %q", machines[0].Name, "sm-1")
	}
}

func TestFetchStateMachines_Pagination(t *testing.T) {
	callCount := 0
	mock := &mockSFNClient{
		listStateMachinesFn: func(ctx context.Context, params *sfn.ListStateMachinesInput, optFns ...func(*sfn.Options)) (*sfn.ListStateMachinesOutput, error) {
			callCount++
			if callCount == 1 {
				return &sfn.ListStateMachinesOutput{
					StateMachines: []types.StateMachineListItem{
						{Name: aws.String("sm-1"), StateMachineArn: aws.String("arn:1"), Type: types.StateMachineTypeStandard},
					},
					NextToken: aws.String("token-1"),
				}, nil
			}
			return &sfn.ListStateMachinesOutput{
				StateMachines: []types.StateMachineListItem{
					{Name: aws.String("sm-2"), StateMachineArn: aws.String("arn:2"), Type: types.StateMachineTypeStandard},
				},
			}, nil
		},
	}
	svc := &Service{Client: mock}
	machines, err := svc.FetchStateMachines(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(machines) != 2 {
		t.Fatalf("expected 2 machines, got %d", len(machines))
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls, got %d", callCount)
	}
}

func TestFetchStateMachines_FilterStandard(t *testing.T) {
	mock := &mockSFNClient{
		listStateMachinesFn: func(ctx context.Context, params *sfn.ListStateMachinesInput, optFns ...func(*sfn.Options)) (*sfn.ListStateMachinesOutput, error) {
			return &sfn.ListStateMachinesOutput{
				StateMachines: []types.StateMachineListItem{
					{Name: aws.String("standard-sm"), StateMachineArn: aws.String("arn:standard"), Type: types.StateMachineTypeStandard},
					{Name: aws.String("express-sm"), StateMachineArn: aws.String("arn:express"), Type: types.StateMachineTypeExpress},
				},
			}, nil
		},
	}
	svc := &Service{Client: mock}
	machines, err := svc.FetchStateMachines(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(machines) != 1 {
		t.Fatalf("expected 1 machine (EXPRESS filtered), got %d", len(machines))
	}
	if machines[0].Name != "standard-sm" {
		t.Errorf("expected 'standard-sm', got %q", machines[0].Name)
	}
}

func TestFetchStateMachines_Empty(t *testing.T) {
	mock := &mockSFNClient{
		listStateMachinesFn: func(ctx context.Context, params *sfn.ListStateMachinesInput, optFns ...func(*sfn.Options)) (*sfn.ListStateMachinesOutput, error) {
			return &sfn.ListStateMachinesOutput{
				StateMachines: []types.StateMachineListItem{},
			}, nil
		},
	}
	svc := &Service{Client: mock}
	machines, err := svc.FetchStateMachines(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(machines) != 0 {
		t.Errorf("expected 0 machines, got %d", len(machines))
	}
}

func TestFetchStateMachines_APIError(t *testing.T) {
	mock := &mockSFNClient{
		listStateMachinesFn: func(ctx context.Context, params *sfn.ListStateMachinesInput, optFns ...func(*sfn.Options)) (*sfn.ListStateMachinesOutput, error) {
			return nil, fmt.Errorf("access denied")
		},
	}
	svc := &Service{Client: mock}
	_, err := svc.FetchStateMachines(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- FetchLatestStatus tests ---

func TestFetchLatestStatus_Success(t *testing.T) {
	mock := &mockSFNClient{
		listExecutionsFn: func(ctx context.Context, params *sfn.ListExecutionsInput, optFns ...func(*sfn.Options)) (*sfn.ListExecutionsOutput, error) {
			if params.MaxResults != 1 {
				t.Errorf("expected maxResults=1, got %d", params.MaxResults)
			}
			return &sfn.ListExecutionsOutput{
				Executions: []types.ExecutionListItem{
					{Status: types.ExecutionStatusSucceeded},
				},
			}, nil
		},
	}
	svc := &Service{Client: mock, MaxConcurrency: 10}
	machines := []StateMachine{
		{Name: "sm-1", ARN: "arn:1"},
		{Name: "sm-2", ARN: "arn:2"},
	}
	result, err := svc.FetchLatestStatus(context.Background(), machines)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, m := range result {
		if m.LatestStatus != string(types.ExecutionStatusSucceeded) {
			t.Errorf("machine %s: expected status SUCCEEDED, got %q", m.Name, m.LatestStatus)
		}
	}
}

func TestFetchLatestStatus_NoExecutions(t *testing.T) {
	mock := &mockSFNClient{
		listExecutionsFn: func(ctx context.Context, params *sfn.ListExecutionsInput, optFns ...func(*sfn.Options)) (*sfn.ListExecutionsOutput, error) {
			return &sfn.ListExecutionsOutput{
				Executions: []types.ExecutionListItem{},
			}, nil
		},
	}
	svc := &Service{Client: mock, MaxConcurrency: 10}
	machines := []StateMachine{
		{Name: "sm-no-exec", ARN: "arn:no-exec"},
	}
	result, err := svc.FetchLatestStatus(context.Background(), machines)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result[0].LatestStatus != "" {
		t.Errorf("expected empty status for machine with no executions, got %q", result[0].LatestStatus)
	}
}

func TestFetchLatestStatus_ConcurrencyLimit(t *testing.T) {
	tracker := &concurrencyTracker{}
	mock := &mockSFNClient{
		listExecutionsFn: func(ctx context.Context, params *sfn.ListExecutionsInput, optFns ...func(*sfn.Options)) (*sfn.ListExecutionsOutput, error) {
			tracker.enter()
			defer tracker.leave()
			// Simulate some work
			time.Sleep(10 * time.Millisecond)
			return &sfn.ListExecutionsOutput{
				Executions: []types.ExecutionListItem{
					{Status: types.ExecutionStatusSucceeded},
				},
			}, nil
		},
	}

	// Create 20 state machines to ensure concurrency limit is exercised
	machines := make([]StateMachine, 20)
	for i := range machines {
		machines[i] = StateMachine{Name: fmt.Sprintf("sm-%d", i), ARN: fmt.Sprintf("arn:%d", i)}
	}

	svc := &Service{Client: mock, MaxConcurrency: 10}
	_, err := svc.FetchLatestStatus(context.Background(), machines)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	maxConcurrent := tracker.max.Load()
	if maxConcurrent > 10 {
		t.Errorf("max concurrency exceeded limit: got %d, want <= 10", maxConcurrent)
	}
}

// --- FetchExecutionHistory tests ---

func TestFetchExecutionHistory_Success(t *testing.T) {
	now := time.Now()
	mock := &mockSFNClient{
		listExecutionsFn: func(ctx context.Context, params *sfn.ListExecutionsInput, optFns ...func(*sfn.Options)) (*sfn.ListExecutionsOutput, error) {
			if params.MaxResults != 20 {
				t.Errorf("expected maxResults=20, got %d", params.MaxResults)
			}
			executions := make([]types.ExecutionListItem, 3)
			for i := range executions {
				executions[i] = types.ExecutionListItem{
					ExecutionArn: aws.String(fmt.Sprintf("arn:exec:%d", i)),
					Name:         aws.String(fmt.Sprintf("exec-%d", i)),
					Status:       types.ExecutionStatusSucceeded,
					StartDate:    aws.Time(now.Add(-time.Duration(i) * time.Hour)),
					StopDate:     aws.Time(now.Add(-time.Duration(i)*time.Hour + 30*time.Minute)),
				}
			}
			return &sfn.ListExecutionsOutput{Executions: executions}, nil
		},
		getExecutionHistoryFn: func(ctx context.Context, params *sfn.GetExecutionHistoryInput, optFns ...func(*sfn.Options)) (*sfn.GetExecutionHistoryOutput, error) {
			return &sfn.GetExecutionHistoryOutput{Events: []types.HistoryEvent{}}, nil
		},
		describeExecutionFn: func(ctx context.Context, params *sfn.DescribeExecutionInput, optFns ...func(*sfn.Options)) (*sfn.DescribeExecutionOutput, error) {
			return &sfn.DescribeExecutionOutput{
				ExecutionArn:    params.ExecutionArn,
				Status:          types.ExecutionStatusSucceeded,
				StateMachineArn: aws.String("arn:sm"),
				StartDate:       aws.Time(now),
				Input:           aws.String(`{"key":"value"}`),
			}, nil
		},
	}
	svc := &Service{Client: mock, MaxConcurrency: 10}
	executions, _, err := svc.FetchExecutionHistory(context.Background(), "arn:sm:test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(executions) != 3 {
		t.Fatalf("expected 3 executions, got %d", len(executions))
	}
}

func TestFetchExecutionHistory_WithFailedState(t *testing.T) {
	now := time.Now()
	mock := &mockSFNClient{
		listExecutionsFn: func(ctx context.Context, params *sfn.ListExecutionsInput, optFns ...func(*sfn.Options)) (*sfn.ListExecutionsOutput, error) {
			return &sfn.ListExecutionsOutput{
				Executions: []types.ExecutionListItem{
					{
						ExecutionArn: aws.String("arn:exec:failed"),
						Name:         aws.String("exec-failed"),
						Status:       types.ExecutionStatusFailed,
						StartDate:    aws.Time(now),
						StopDate:     aws.Time(now.Add(5 * time.Minute)),
					},
				},
			}, nil
		},
		getExecutionHistoryFn: func(ctx context.Context, params *sfn.GetExecutionHistoryInput, optFns ...func(*sfn.Options)) (*sfn.GetExecutionHistoryOutput, error) {
			if !params.ReverseOrder {
				t.Error("expected ReverseOrder=true for failed execution history")
			}
			return &sfn.GetExecutionHistoryOutput{
				Events: []types.HistoryEvent{
					{
						Id:              3,
						Type:            types.HistoryEventTypeTaskFailed,
						PreviousEventId: 2,
						TaskFailedEventDetails: &types.TaskFailedEventDetails{
							Resource:     aws.String("lambda"),
							ResourceType: aws.String("lambda"),
							Error:        aws.String("SomeError"),
						},
					},
					{
						Id:   2,
						Type: types.HistoryEventTypeTaskStateEntered,
						StateEnteredEventDetails: &types.StateEnteredEventDetails{
							Name: aws.String("ProcessPayment"),
						},
					},
				},
			}, nil
		},
		describeExecutionFn: func(ctx context.Context, params *sfn.DescribeExecutionInput, optFns ...func(*sfn.Options)) (*sfn.DescribeExecutionOutput, error) {
			return &sfn.DescribeExecutionOutput{
				ExecutionArn:    params.ExecutionArn,
				Status:          types.ExecutionStatusFailed,
				StateMachineArn: aws.String("arn:sm"),
				StartDate:       aws.Time(now),
				Input:           aws.String(`{}`),
			}, nil
		},
	}
	svc := &Service{Client: mock, MaxConcurrency: 10}
	executions, _, err := svc.FetchExecutionHistory(context.Background(), "arn:sm:test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(executions) != 1 {
		t.Fatalf("expected 1 execution, got %d", len(executions))
	}
	if executions[0].FailedState != "ProcessPayment" {
		t.Errorf("expected FailedState 'ProcessPayment', got %q", executions[0].FailedState)
	}
}

func TestFetchExecutionHistory_FailedStateTraversal(t *testing.T) {
	// Tests that previousEventId is followed to find the state name
	now := time.Now()
	mock := &mockSFNClient{
		listExecutionsFn: func(ctx context.Context, params *sfn.ListExecutionsInput, optFns ...func(*sfn.Options)) (*sfn.ListExecutionsOutput, error) {
			return &sfn.ListExecutionsOutput{
				Executions: []types.ExecutionListItem{
					{
						ExecutionArn: aws.String("arn:exec:failed"),
						Name:         aws.String("exec-failed"),
						Status:       types.ExecutionStatusFailed,
						StartDate:    aws.Time(now),
						StopDate:     aws.Time(now.Add(5 * time.Minute)),
					},
				},
			}, nil
		},
		getExecutionHistoryFn: func(ctx context.Context, params *sfn.GetExecutionHistoryInput, optFns ...func(*sfn.Options)) (*sfn.GetExecutionHistoryOutput, error) {
			// Simulate a chain: LambdaFunctionFailed -> TaskScheduled -> TaskStateEntered
			return &sfn.GetExecutionHistoryOutput{
				Events: []types.HistoryEvent{
					{
						Id:              5,
						Type:            types.HistoryEventTypeLambdaFunctionFailed,
						PreviousEventId: 4,
						LambdaFunctionFailedEventDetails: &types.LambdaFunctionFailedEventDetails{
							Error: aws.String("Lambda.TimeoutError"),
						},
					},
					{
						Id:              4,
						Type:            types.HistoryEventTypeLambdaFunctionStarted,
						PreviousEventId: 3,
					},
					{
						Id:              3,
						Type:            types.HistoryEventTypeLambdaFunctionScheduled,
						PreviousEventId: 2,
					},
					{
						Id:   2,
						Type: types.HistoryEventTypeTaskStateEntered,
						StateEnteredEventDetails: &types.StateEnteredEventDetails{
							Name: aws.String("InvokeLambda"),
						},
					},
				},
			}, nil
		},
		describeExecutionFn: func(ctx context.Context, params *sfn.DescribeExecutionInput, optFns ...func(*sfn.Options)) (*sfn.DescribeExecutionOutput, error) {
			return &sfn.DescribeExecutionOutput{
				ExecutionArn:    params.ExecutionArn,
				Status:          types.ExecutionStatusFailed,
				StateMachineArn: aws.String("arn:sm"),
				StartDate:       aws.Time(now),
				Input:           aws.String(`{}`),
			}, nil
		},
	}
	svc := &Service{Client: mock, MaxConcurrency: 10}
	executions, _, err := svc.FetchExecutionHistory(context.Background(), "arn:sm:test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(executions) != 1 {
		t.Fatalf("expected 1 execution, got %d", len(executions))
	}
	if executions[0].FailedState != "InvokeLambda" {
		t.Errorf("expected FailedState 'InvokeLambda', got %q", executions[0].FailedState)
	}
}

func TestFetchExecutionHistory_WithInputParam(t *testing.T) {
	now := time.Now()
	expectedInput := `{"orderId":"12345","amount":100}`
	mock := &mockSFNClient{
		listExecutionsFn: func(ctx context.Context, params *sfn.ListExecutionsInput, optFns ...func(*sfn.Options)) (*sfn.ListExecutionsOutput, error) {
			return &sfn.ListExecutionsOutput{
				Executions: []types.ExecutionListItem{
					{
						ExecutionArn: aws.String("arn:exec:1"),
						Name:         aws.String("exec-1"),
						Status:       types.ExecutionStatusSucceeded,
						StartDate:    aws.Time(now),
						StopDate:     aws.Time(now.Add(10 * time.Minute)),
					},
				},
			}, nil
		},
		getExecutionHistoryFn: func(ctx context.Context, params *sfn.GetExecutionHistoryInput, optFns ...func(*sfn.Options)) (*sfn.GetExecutionHistoryOutput, error) {
			return &sfn.GetExecutionHistoryOutput{Events: []types.HistoryEvent{}}, nil
		},
		describeExecutionFn: func(ctx context.Context, params *sfn.DescribeExecutionInput, optFns ...func(*sfn.Options)) (*sfn.DescribeExecutionOutput, error) {
			return &sfn.DescribeExecutionOutput{
				ExecutionArn:    params.ExecutionArn,
				Status:          types.ExecutionStatusSucceeded,
				StateMachineArn: aws.String("arn:sm"),
				StartDate:       aws.Time(now),
				Input:           aws.String(expectedInput),
			}, nil
		},
	}
	svc := &Service{Client: mock, MaxConcurrency: 10}
	executions, _, err := svc.FetchExecutionHistory(context.Background(), "arn:sm:test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(executions) != 1 {
		t.Fatalf("expected 1 execution, got %d", len(executions))
	}
	if executions[0].InputParam != expectedInput {
		t.Errorf("expected InputParam %q, got %q", expectedInput, executions[0].InputParam)
	}
}

func TestFetchExecutionHistory_APIError(t *testing.T) {
	mock := &mockSFNClient{
		listExecutionsFn: func(ctx context.Context, params *sfn.ListExecutionsInput, optFns ...func(*sfn.Options)) (*sfn.ListExecutionsOutput, error) {
			return nil, fmt.Errorf("throttling exception")
		},
	}
	svc := &Service{Client: mock, MaxConcurrency: 10}
	_, _, err := svc.FetchExecutionHistory(context.Background(), "arn:sm:test", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
