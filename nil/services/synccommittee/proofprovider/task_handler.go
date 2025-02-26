package proofprovider

import (
	"context"

	"github.com/NilFoundation/nil/nil/common"
	"github.com/NilFoundation/nil/nil/services/synccommittee/internal/api"
	"github.com/NilFoundation/nil/nil/services/synccommittee/internal/log"
	"github.com/NilFoundation/nil/nil/services/synccommittee/internal/types"
	"github.com/rs/zerolog"
)

type HandlerTaskStorage interface {
	AddTaskEntries(ctx context.Context, tasks ...*types.TaskEntry) error
}

type taskHandler struct {
	taskStorage HandlerTaskStorage
	resultSaver TaskResultSaver
	skipRate    int
	taskNum     int
	timer       common.Timer
	logger      zerolog.Logger
}

func newTaskHandler(
	taskStorage HandlerTaskStorage, resultSaver TaskResultSaver, skipRate int, timer common.Timer, logger zerolog.Logger,
) api.TaskHandler {
	return &taskHandler{
		taskStorage: taskStorage,
		resultSaver: resultSaver,
		skipRate:    skipRate,
		taskNum:     0,
		timer:       timer,
		logger:      logger,
	}
}

func (h *taskHandler) Handle(ctx context.Context, executorId types.TaskExecutorId, task *types.Task) error {
	if (task.TaskType != types.ProofBlock) && (task.TaskType != types.AggregateProofs) {
		return types.NewTaskErrNotSupportedType(task.TaskType)
	}
	// skip task
	taskIdx := h.taskNum % 10
	h.taskNum++
	if taskIdx < h.skipRate {
		log.NewTaskEvent(h.logger, zerolog.InfoLevel, task).Msg("Skip task")
		skippedTaskResult := types.NewSuccessProviderTaskResult(
			task.Id,
			executorId,
			types.TaskOutputArtifacts{},
			task.BlockHash.Bytes(),
		)
		err := h.resultSaver.Put(ctx, skippedTaskResult)
		if err != nil {
			log.NewTaskEvent(h.logger, zerolog.ErrorLevel, task).
				Err(err).
				Msgf("Failed to send skipped task result")
		}
		return nil
	}

	log.NewTaskEvent(h.logger, zerolog.InfoLevel, task).Msg("Creating proof tasks for block")

	var err error
	if task.TaskType == types.ProofBlock {
		blockTasks := h.prepareTasksForBlock(task)

		for _, taskEntry := range blockTasks {
			taskEntry.Task.ParentTaskId = &task.Id
		}

		err = h.taskStorage.AddTaskEntries(ctx, blockTasks...)
	} else {
		currentTime := h.timer.NowTime()
		aggregateTaskEntry := task.AsNewChildEntry(currentTime)
		err = h.taskStorage.AddTaskEntries(ctx, aggregateTaskEntry)
	}

	if err != nil {
		log.NewTaskEvent(h.logger, zerolog.ErrorLevel, task).Err(err).Msg("Failed to create proof task")
	} else {
		log.NewTaskEvent(h.logger, zerolog.InfoLevel, task).Msg("Proof tasks created")
	}
	return err
}

func (h *taskHandler) prepareTasksForBlock(providerTask *types.Task) []*types.TaskEntry {
	currentTime := h.timer.NowTime()
	taskEntries := make([]*types.TaskEntry, 0)

	// Final task, depends on partial proofs, aggregate FRI and consistency checks
	mergeProofTaskEntry := types.NewMergeProofTaskEntry(providerTask, currentTime)
	taskEntries = append(taskEntries, mergeProofTaskEntry)

	// Third level of circuit-dependent tasks
	consistencyCheckTasks := make(map[types.CircuitType]*types.TaskEntry)
	for ct := range types.Circuits() {
		checkTaskEntry := types.NewFRIConsistencyCheckTaskEntry(providerTask, ct, currentTime)
		taskEntries = append(taskEntries, checkTaskEntry)
		consistencyCheckTasks[ct] = checkTaskEntry

		// FRI consistency check task result goes to merge proof task
		mergeProofTaskEntry.AddDependency(checkTaskEntry)
	}

	// aggregate FRI task depends on all the following tasks
	aggFRITaskEntry := types.NewAggregateFRITaskEntry(providerTask, currentTime)
	taskEntries = append(taskEntries, aggFRITaskEntry)
	// Aggregate FRI task result must be forwarded to merge proof task
	mergeProofTaskEntry.AddDependency(aggFRITaskEntry)

	for _, checkTaskEntry := range consistencyCheckTasks {
		// Also aggregate FRI task result goes to all consistency check tasks
		checkTaskEntry.AddDependency(aggFRITaskEntry)
	}

	// Second level of circuit-dependent tasks
	combinedQTasks := make(map[types.CircuitType]*types.TaskEntry)
	for ct := range types.Circuits() {
		combinedQTaskEntry := types.NewCombinedQTaskEntry(providerTask, ct, currentTime)
		taskEntries = append(taskEntries, combinedQTaskEntry)
		combinedQTasks[ct] = combinedQTaskEntry
	}

	for ct, combQEntry := range combinedQTasks {
		// Combined Q task result goes to aggregate FRI task and consistency check task
		aggFRITaskEntry.AddDependency(combQEntry)
		consistencyCheckTasks[ct].AddDependency(combQEntry)
	}

	// aggregate challenge task depends on all the following tasks
	aggChallengeTaskEntry := types.NewAggregateChallengeTaskEntry(providerTask, currentTime)
	taskEntries = append(taskEntries, aggChallengeTaskEntry)

	// aggregate challenges task result goes to all combined Q tasks
	for _, combQEntry := range combinedQTasks {
		combQEntry.AddDependency(aggChallengeTaskEntry)
	}

	// One more destination of aggregate challenge task result is aggregate FRI task
	aggFRITaskEntry.AddDependency(aggChallengeTaskEntry)

	// Create partial proof tasks (bottom level, no dependencies)
	for ct := range types.Circuits() {
		partialProveTaskEntry := types.NewPartialProveTaskEntry(providerTask, ct, currentTime)
		taskEntries = append(taskEntries, partialProveTaskEntry)

		// Partial proof results go to all other levels of tasks
		aggChallengeTaskEntry.AddDependency(partialProveTaskEntry)
		combinedQTasks[ct].AddDependency(partialProveTaskEntry)
		aggFRITaskEntry.AddDependency(partialProveTaskEntry)
		consistencyCheckTasks[ct].AddDependency(partialProveTaskEntry)
		mergeProofTaskEntry.AddDependency(partialProveTaskEntry)
	}

	return taskEntries
}
