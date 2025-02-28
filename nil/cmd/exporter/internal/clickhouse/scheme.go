package clickhouse

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/NilFoundation/nil/nil/common"
	"github.com/NilFoundation/nil/nil/common/check"
)

var tableSchemeCache map[string]reflectedScheme = nil

func initSchemeCache() map[string]reflectedScheme {
	tableScheme := make(map[string]reflectedScheme)

	blockScheme, err := reflectSchemeToClickhouse(&BlockWithBinary{})
	check.PanicIfErr(err)

	tableScheme["blocks"] = blockScheme
	transactionScheme, err := reflectSchemeToClickhouse(&TransactionWithBinary{})
	check.PanicIfErr(err)

	tableScheme["transactions"] = transactionScheme
	logScheme, err := reflectSchemeToClickhouse(&LogWithBinary{})
	check.PanicIfErr(err)

	tableScheme["logs"] = logScheme

	return tableScheme
}

func getTableScheme() map[string]reflectedScheme {
	if tableSchemeCache == nil {
		tableSchemeCache = initSchemeCache()
	}

	return tableSchemeCache
}

func setupScheme(ctx context.Context, conn driver.Conn, tableName string, keys []string) error {
	tableScheme := getTableScheme()
	scheme, ok := tableScheme[tableName]
	if !ok {
		return fmt.Errorf("scheme for %s not found", tableName)
	}

	query := scheme.CreateTableQuery(tableName, "ReplacingMergeTree", keys, keys)
	if err := conn.Exec(ctx, query); err != nil {
		return fmt.Errorf("failed to create table %s: %w", tableName, err)
	}

	return nil
}

func setupSchemes(ctx context.Context, conn driver.Conn) error {
	if err := setupScheme(ctx, conn,
		"blocks", []string{"shard_id", "hash"}); err != nil {
		return err
	}

	if err := setupScheme(ctx, conn,
		"transactions", []string{"hash", "outgoing"}); err != nil {
		return err
	}

	if err := setupScheme(ctx, conn,
		"logs", []string{"transaction_hash"}); err != nil {
		return err
	}

	return nil
}

func createTableQuery(tableName, fields, engine string, primaryKeys, orderKeys []string) string {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s
		(%s)
		ENGINE = %s
		PRIMARY KEY (%s)
		ORDER BY (%s)
`, tableName, fields, engine, strings.Join(primaryKeys, ", "), strings.Join(orderKeys, ", "))
	logger.Debug().Msgf("CreateTableQuery: %s", query)
	return query
}

func tableExists(ctx context.Context, conn driver.Conn, tableName string) (bool, error) {
	var count uint64
	if err := conn.QueryRow(ctx, `
		SELECT count()
		FROM system.tables
		WHERE database = currentDatabase() AND name = $1
	`, tableName).Scan(&count); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return count > 0, nil
}

func readVersion(ctx context.Context, conn driver.Conn) (common.Hash, error) {
	if exists, err := tableExists(ctx, conn, "blocks"); err != nil || !exists {
		return common.Hash{}, err
	}

	var version common.Hash
	if err := conn.QueryRow(ctx, `
		SELECT hash
		FROM blocks
		WHERE shard_id = 0 AND id = 0
	`).Scan(&version); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return common.Hash{}, nil
		}
		return common.Hash{}, err
	}

	return version, nil
}
