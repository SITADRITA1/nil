package tracer

import (
	"context"
	"errors"
	"math/big"

	"github.com/NilFoundation/nil/nil/common"
	"github.com/NilFoundation/nil/nil/common/logging"
	"github.com/NilFoundation/nil/nil/internal/db"
	"github.com/NilFoundation/nil/nil/internal/execution"
	"github.com/NilFoundation/nil/nil/internal/types"
	"github.com/NilFoundation/nil/nil/internal/vm"
	"github.com/NilFoundation/nil/nil/services/rpc/jsonrpc"
	"github.com/NilFoundation/nil/nil/services/rpc/transport"
	"github.com/NilFoundation/nil/nil/services/synccommittee/prover/tracer/api"
	"github.com/rs/zerolog"
)

type RemoteTracer interface {
	GetBlockTraces(ctx context.Context, aggTraces ExecutionTraces, shardId types.ShardId, blockRef transport.BlockReference) error
}

type RemoteTracerImpl struct {
	client api.RpcClient
	logger zerolog.Logger
}

var _ RemoteTracer = new(RemoteTracerImpl)

type TraceConfig struct {
	ShardID      types.ShardId
	BlockIDs     []transport.BlockReference
	BaseFileName string
	MarshalMode  MarshalMode
}

func NewRemoteTracer(client api.RpcClient, logger zerolog.Logger) (*RemoteTracerImpl, error) {
	return &RemoteTracerImpl{
		client: client,
		logger: logger,
	}, nil
}

func (rt *RemoteTracerImpl) GetBlockTraces(
	ctx context.Context,
	aggTraces ExecutionTraces,
	shardId types.ShardId,
	blockRef transport.BlockReference,
) error {
	dbgBlock, err := rt.client.GetDebugBlock(ctx, shardId, blockRef, true)
	if err != nil {
		return err
	}
	if dbgBlock == nil {
		return errors.New("client returned nil block")
	}
	decodedDbgBlock, err := dbgBlock.DecodeSSZ()
	if err != nil {
		return err
	}
	if decodedDbgBlock.Id == 0 {
		// TODO: prove genesis block generation?
		return ErrCantProofGenesisBlock
	}

	prevBlock, err := rt.client.GetBlock(ctx, shardId, transport.BlockNumber(decodedDbgBlock.Id-1), true)
	if err != nil {
		return err
	}
	if prevBlock == nil {
		return errors.New("client returned nil block")
	}

	getHashFunc := func(blkNum uint64) (common.Hash, error) {
		// TODO: try to replace it with prevBlock.Hash
		_ = prevBlock.Hash

		return decodedDbgBlock.Hash(shardId), nil
	}

	blkContext := &vm.BlockContext{
		GetHash:     getHashFunc,
		BlockNumber: decodedDbgBlock.Id.Uint64(),
		Random:      &common.EmptyHash,
		BaseFee:     decodedDbgBlock.BaseFee.ToBig(),
		// TODO: adjust when `NewEVMBlockContext` uses non-hardcoded 10 value.
		// Seems like we need to include this into API Block response.
		BlobBaseFee: big.NewInt(10),
		Time:        decodedDbgBlock.Timestamp,
	}

	chainConfig, err := rt.getConfigForBlock(ctx, decodedDbgBlock.Block, shardId)
	if err != nil {
		return err
	}

	localDb, err := db.NewBadgerDbInMemory()
	if err != nil {
		return err
	}

	stateDB, err := NewTracerStateDB(
		ctx,
		aggTraces,
		rt.client,
		shardId,
		prevBlock.Number,
		blkContext,
		localDb,
		chainConfig,
		rt.logger,
	)
	if err != nil {
		return err
	}

	ecdsaTracer := NewEcdsaTracer()

	for _, inTxn := range decodedDbgBlock.InTransactions {
		_, txnHadErr := decodedDbgBlock.Errors[inTxn.Hash()]
		if txnHadErr {
			continue
		}

		if err := ecdsaTracer.TraceTx(inTxn); err != nil {
			return err
		}

		if inTxn.Flags.GetBit(types.TransactionFlagResponse) {
			return errors.New("can't process responses in prover, refer to TryProcessResponse of ExecutionState")
		}

		stateDB.AddInTransaction(inTxn)
		if err := stateDB.HandleInTransaction(inTxn, execution.NewTransactionPayer(inTxn, stateDB)); err != nil { //nolint:contextcheck
			return err
		}
	}

	err = stateDB.FinalizeTraces()
	if err != nil {
		return err
	}
	aggTraces.AddEcdsaSigns(ecdsaTracer.Finalize())

	// Print stats
	stats := stateDB.Stats
	rt.logger.Info().
		Uint("processedInTransactions", stats.ProcessedInTxnsN).
		Uint("totalInTransactions", uint(len(decodedDbgBlock.InTransactions))).
		Uint("operations", stats.OpsN).
		Uint("stackOperations", stats.StackOpsN).
		Uint("memoryOperations", stats.MemoryOpsN).
		Uint("stateOperations", stats.StateOpsN).
		Uint("copyOperations", stats.CopyOpsN).
		Uint("expOperations", stats.ExpOpsN).
		Uint("keccakOperations", stats.KeccakOpsN).
		Uint("affectedContracts", stats.AffectedContractsN).
		Msg("Tracer stats")

	return nil
}

func GenerateTrace(ctx context.Context, rpcClient api.RpcClient, cfg *TraceConfig) error {
	remoteTracer, err := NewRemoteTracer(rpcClient, logging.NewLogger("tracer"))
	if err != nil {
		return err
	}
	aggTraces := NewExecutionTraces()
	for _, blockID := range cfg.BlockIDs {
		err := remoteTracer.GetBlockTraces(ctx, aggTraces, cfg.ShardID, blockID)
		if err != nil {
			return err
		}
	}

	return SerializeToFile(aggTraces, cfg.MarshalMode, cfg.BaseFileName)
}

func (rt *RemoteTracerImpl) getConfigForBlock(ctx context.Context, block *types.Block, shardId types.ShardId) (*jsonrpc.ChainConfig, error) {
	var blockWithConfig common.Hash
	if shardId.IsMainShard() {
		// fetch config from prev main chain block
		blockWithConfig = block.PrevBlock
	} else {
		// fetch config from corresponding main chain hash
		blockWithConfig = block.MainChainHash
	}
	dbgBlock, err := rt.client.GetDebugBlock(ctx, shardId, blockWithConfig, true)
	if err != nil {
		return nil, err
	}
	if dbgBlock == nil {
		return nil, errors.New("client returned nil block")
	}

	return dbgBlock.Config, nil
}
