package core

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/NilFoundation/nil/nil/client"
	"github.com/NilFoundation/nil/nil/common"
	"github.com/NilFoundation/nil/nil/common/concurrent"
	"github.com/NilFoundation/nil/nil/common/logging"
	coreTypes "github.com/NilFoundation/nil/nil/internal/types"
	"github.com/NilFoundation/nil/nil/services/rpc/jsonrpc"
	"github.com/NilFoundation/nil/nil/services/synccommittee/core/batches"
	"github.com/NilFoundation/nil/nil/services/synccommittee/core/batches/blob"
	v1 "github.com/NilFoundation/nil/nil/services/synccommittee/core/batches/encode/v1"
	"github.com/NilFoundation/nil/nil/services/synccommittee/internal/metrics"
	"github.com/NilFoundation/nil/nil/services/synccommittee/internal/srv"
	"github.com/NilFoundation/nil/nil/services/synccommittee/internal/types"
	"github.com/rs/zerolog"
)

type AggregatorMetrics interface {
	metrics.BasicMetrics
	RecordMainBlockFetched(ctx context.Context)
	RecordBlockBatchSize(ctx context.Context, batchSize int64)
}

type AggregatorTaskStorage interface {
	AddTaskEntries(ctx context.Context, tasks ...*types.TaskEntry) error
}

type AggregatorBlockStorage interface {
	TryGetLatestFetched(ctx context.Context) (*types.MainBlockRef, error)
	TryGetLatestBatchId(ctx context.Context) (*types.BatchId, error)
	SetBlockBatch(ctx context.Context, batch *types.BlockBatch) error
}

type aggregator struct {
	logger         zerolog.Logger
	rpcClient      client.Client
	blockStorage   AggregatorBlockStorage
	taskStorage    AggregatorTaskStorage
	batchCommitter batches.BatchCommitter
	timer          common.Timer
	metrics        AggregatorMetrics
	workerAction   *concurrent.Suspendable
}

func NewAggregator(
	rpcClient client.Client,
	blockStorage AggregatorBlockStorage,
	taskStorage AggregatorTaskStorage,
	timer common.Timer,
	logger zerolog.Logger,
	metrics AggregatorMetrics,
	pollingDelay time.Duration,
) *aggregator {
	agg := &aggregator{
		rpcClient:    rpcClient,
		blockStorage: blockStorage,
		taskStorage:  taskStorage,
		batchCommitter: batches.NewBatchCommitter(
			v1.NewEncoder(logger),
			blob.NewBuilder(),
			nil, // TODO
			logger,
			batches.DefaultCommitOptions(),
		),
		timer:   timer,
		metrics: metrics,
	}

	agg.workerAction = concurrent.NewSuspendable(agg.runIteration, pollingDelay)
	agg.logger = srv.WorkerLogger(logger, agg)
	return agg
}

func (agg *aggregator) Name() string {
	return "aggregator"
}

func (agg *aggregator) Run(ctx context.Context, started chan<- struct{}) error {
	agg.logger.Info().Msg("starting blocks fetching")

	err := agg.workerAction.Run(ctx, started)

	if err == nil || errors.Is(err, context.Canceled) {
		agg.logger.Info().Msg("blocks fetching stopped")
	} else {
		agg.logger.Error().Err(err).Msg("error running aggregator, stopped")
	}

	return err
}

func (agg *aggregator) Pause(ctx context.Context) error {
	paused, err := agg.workerAction.Pause(ctx)
	if err != nil {
		return err
	}
	if paused {
		agg.logger.Info().Msg("blocks fetching paused")
	}
	return nil
}

func (agg *aggregator) Resume(ctx context.Context) error {
	resumed, err := agg.workerAction.Resume(ctx)
	if err != nil {
		return err
	}
	if resumed {
		agg.logger.Info().Msg("blocks fetching resumed")
	}
	return nil
}

func (agg *aggregator) runIteration(ctx context.Context) {
	err := agg.processNewBlocks(ctx)

	if errors.Is(err, types.ErrBatchNotReady) {
		agg.logger.Warn().Err(err).Msg("received unready block batch, skipping")
		return
	}

	if err != nil {
		agg.logger.Error().Err(err).Msg("error during processing new blocks")
		agg.metrics.RecordError(ctx, agg.Name())
	}
}

// processNewBlocks fetches and processes new blocks for all shards.
// It handles the overall flow of block synchronization and proof creation.
func (agg *aggregator) processNewBlocks(ctx context.Context) error {
	latestBlock, err := agg.fetchLatestBlockRef(ctx)
	if err != nil {
		return err
	}

	if err := agg.processShardBlocks(ctx, *latestBlock); err != nil {
		// todo: launch block re-fetching in case of ErrBlockMismatch
		return fmt.Errorf("error processing blocks: %w", err)
	}

	return nil
}

// fetchLatestBlocks retrieves the latest block for main shard
func (agg *aggregator) fetchLatestBlockRef(ctx context.Context) (*types.MainBlockRef, error) {
	block, err := agg.rpcClient.GetBlock(ctx, coreTypes.MainShardId, "latest", false)
	if err != nil {
		return nil, fmt.Errorf("error fetching latest block from shard %d: %w", coreTypes.MainShardId, err)
	}
	return types.NewBlockRef(block)
}

// processShardBlocks handles the processing of new blocks for the main shard.
// It fetches new blocks, updates the storage, and records relevant metrics.
func (agg *aggregator) processShardBlocks(ctx context.Context, actualLatest types.MainBlockRef) error {
	latestFetched, err := agg.blockStorage.TryGetLatestFetched(ctx)
	if err != nil {
		return fmt.Errorf("error reading latest fetched block for the main shard: %w", err)
	}

	fetchingRange, err := types.GetBlocksFetchingRange(latestFetched, actualLatest)
	if err != nil {
		return err
	}

	if fetchingRange == nil {
		agg.logger.Debug().
			Stringer(logging.FieldShardId, coreTypes.MainShardId).
			Stringer(logging.FieldBlockNumber, actualLatest.Number).
			Msg("no new blocks to fetch")
	} else {
		if err := agg.fetchAndProcessBlocks(ctx, *fetchingRange); err != nil {
			return fmt.Errorf("%w: %w", types.ErrBlockProcessing, err)
		}
	}

	return nil
}

// fetchAndProcessBlocks retrieves a range of blocks for a main shard, stores them, creates proof tasks
func (agg *aggregator) fetchAndProcessBlocks(ctx context.Context, blocksRange types.BlocksRange) error {
	shardId := coreTypes.MainShardId
	const requestBatchSize = 20
	results, err := agg.rpcClient.GetBlocksRange(ctx, shardId, blocksRange.Start, blocksRange.End+1, true, requestBatchSize)
	if err != nil {
		return fmt.Errorf("error fetching blocks from shard %d: %w", shardId, err)
	}

	for _, mainShardBlock := range results {
		blockBatch, err := agg.createBlockBatch(ctx, mainShardBlock)
		if err != nil {
			return fmt.Errorf("error creating batch, mainHash=%s: %w", mainShardBlock.Hash, err)
		}

		if err := agg.handleBlockBatch(ctx, blockBatch); err != nil {
			return fmt.Errorf("error handing batch, mainHash=%s: %w", mainShardBlock.Hash, err)
		}
	}

	fetchedLen := int64(len(results))
	agg.logger.Debug().Int64("blkCount", fetchedLen).Stringer(logging.FieldShardId, shardId).Msg("fetched main shard blocks")
	agg.metrics.RecordBlockBatchSize(ctx, fetchedLen)
	return nil
}

func (agg *aggregator) createBlockBatch(ctx context.Context, mainShardBlock *jsonrpc.RPCBlock) (*types.BlockBatch, error) {
	childIds, err := types.ChildBlockIds(mainShardBlock)
	if err != nil {
		return nil, err
	}

	childBlocks := make([]*jsonrpc.RPCBlock, 0, len(childIds))

	for _, childId := range childIds {
		childBlock, err := agg.rpcClient.GetBlock(ctx, childId.ShardId, childId.Hash, true)
		if err != nil {
			return nil, fmt.Errorf(
				"error fetching child block with id=%s, mainHash=%s: %w", childId, mainShardBlock.Hash, err,
			)
		}
		childBlocks = append(childBlocks, childBlock)
	}

	latestBatchId, err := agg.blockStorage.TryGetLatestBatchId(ctx)
	if err != nil {
		return nil, fmt.Errorf("error reading latest batch id: %w", err)
	}

	return types.NewBlockBatch(latestBatchId, mainShardBlock, childBlocks)
}

// handleBlockBatch checks the validity of a block and stores it if valid.
func (agg *aggregator) handleBlockBatch(ctx context.Context, batch *types.BlockBatch) error {
	latestFetched, err := agg.blockStorage.TryGetLatestFetched(ctx)
	if err != nil {
		return fmt.Errorf("error reading latest fetched block from storage: %w", err)
	}
	if err := latestFetched.ValidateChild(batch.MainShardBlock); err != nil {
		return err
	}

	prunedBatch := types.NewPrunedBatch(batch)
	if err := agg.batchCommitter.Commit(ctx, prunedBatch); err != nil {
		return err
	}

	if err := agg.createProofTasks(ctx, batch); err != nil {
		return fmt.Errorf("error creating proof tasks, mainHash=%s: %w", batch.MainShardBlock.Hash, err)
	}

	if err := agg.blockStorage.SetBlockBatch(ctx, batch); err != nil {
		return fmt.Errorf("error storing block batch, mainHash=%s: %w", batch.MainShardBlock.Hash, err)
	}

	agg.metrics.RecordMainBlockFetched(ctx)
	return nil
}

// createProofTask generates proof tasks for block batch
func (agg *aggregator) createProofTasks(ctx context.Context, batch *types.BlockBatch) error {
	currentTime := agg.timer.NowTime()
	proofTasks, err := batch.CreateProofTasks(currentTime)
	if err != nil {
		return fmt.Errorf("error creating proof tasks, mainHash=%s: %w", batch.MainShardBlock.Hash, err)
	}

	if err := agg.taskStorage.AddTaskEntries(ctx, proofTasks...); err != nil {
		return fmt.Errorf("error adding task entries, mainHash=%s: %w", batch.MainShardBlock.Hash, err)
	}

	agg.logger.Debug().
		Stringer(logging.FieldBatchId, batch.Id).
		Msgf("created %d proof tasks, mainHash=%s", len(proofTasks), batch.MainShardBlock.Hash)

	return nil
}
