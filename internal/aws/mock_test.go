package aws

import (
	"context"
	"sync/atomic"

	"github.com/aws/aws-sdk-go-v2/service/sfn"
)

type mockSFNClient struct {
	listStateMachinesFn    func(ctx context.Context, params *sfn.ListStateMachinesInput, optFns ...func(*sfn.Options)) (*sfn.ListStateMachinesOutput, error)
	listExecutionsFn       func(ctx context.Context, params *sfn.ListExecutionsInput, optFns ...func(*sfn.Options)) (*sfn.ListExecutionsOutput, error)
	getExecutionHistoryFn  func(ctx context.Context, params *sfn.GetExecutionHistoryInput, optFns ...func(*sfn.Options)) (*sfn.GetExecutionHistoryOutput, error)
	describeExecutionFn    func(ctx context.Context, params *sfn.DescribeExecutionInput, optFns ...func(*sfn.Options)) (*sfn.DescribeExecutionOutput, error)
}

func (m *mockSFNClient) ListStateMachines(ctx context.Context, params *sfn.ListStateMachinesInput, optFns ...func(*sfn.Options)) (*sfn.ListStateMachinesOutput, error) {
	return m.listStateMachinesFn(ctx, params, optFns...)
}

func (m *mockSFNClient) ListExecutions(ctx context.Context, params *sfn.ListExecutionsInput, optFns ...func(*sfn.Options)) (*sfn.ListExecutionsOutput, error) {
	return m.listExecutionsFn(ctx, params, optFns...)
}

func (m *mockSFNClient) GetExecutionHistory(ctx context.Context, params *sfn.GetExecutionHistoryInput, optFns ...func(*sfn.Options)) (*sfn.GetExecutionHistoryOutput, error) {
	return m.getExecutionHistoryFn(ctx, params, optFns...)
}

func (m *mockSFNClient) DescribeExecution(ctx context.Context, params *sfn.DescribeExecutionInput, optFns ...func(*sfn.Options)) (*sfn.DescribeExecutionOutput, error) {
	return m.describeExecutionFn(ctx, params, optFns...)
}

// concurrencyTracker tracks the maximum number of concurrent calls using atomic operations.
type concurrencyTracker struct {
	current atomic.Int64
	max     atomic.Int64
}

func (ct *concurrencyTracker) enter() {
	cur := ct.current.Add(1)
	for {
		old := ct.max.Load()
		if cur <= old {
			break
		}
		if ct.max.CompareAndSwap(old, cur) {
			break
		}
	}
}

func (ct *concurrencyTracker) leave() {
	ct.current.Add(-1)
}
