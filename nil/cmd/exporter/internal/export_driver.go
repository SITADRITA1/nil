package internal

import (
	"context"

	"github.com/NilFoundation/nil/nil/common"
	"github.com/NilFoundation/nil/nil/internal/types"
)

type BlockWithShardId struct {
	*types.BlockWithExtractedData
	ShardId types.ShardId
}

type SetupParams struct {
	AllowDbDrop bool
	// Version is the hash of the genesis block of the main shard (must become more complex later).
	Version common.Hash
}

type ExportDriver interface {
	SetupScheme(ctx context.Context, params SetupParams) error
	ExportBlocks(context.Context, []*BlockWithShardId) error
	HaveBlock(context.Context, types.ShardId, types.BlockNumber) (bool, error)
	FetchLatestProcessedBlockId(context.Context, types.ShardId) (types.BlockNumber, error)
	FetchEarliestAbsentBlockId(context.Context, types.ShardId) (types.BlockNumber, error)
	FetchNextPresentBlockId(context.Context, types.ShardId, types.BlockNumber) (types.BlockNumber, error)
	Reconnect() error
}
