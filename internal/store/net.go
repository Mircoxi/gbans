package store

import (
	"context"
	"fmt"
	"net"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/leighmacdonald/gbans/pkg/ip2location"
	"github.com/leighmacdonald/steamid/v3/steamid"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// GetBanNetByAddress returns the BanCIDR matching intersecting the supplied ip.
//
// Note that this function does not currently limit results returned. This may change in the future, do not
// rely on this functionality.
func (db *Store) GetBanNetByAddress(ctx context.Context, ipAddr net.IP) ([]BanCIDR, error) {
	const query = `
		SELECT net_id, cidr, origin, created_on, updated_on, reason, reason_text, valid_until, deleted, 
		       note, unban_reason_text, is_enabled, target_id, source_id, appeal_state
		FROM ban_net
		WHERE $1 <<= cidr AND deleted = false AND is_enabled = true`

	var nets []BanCIDR

	rows, errQuery := db.Query(ctx, query, ipAddr.String())
	if errQuery != nil {
		return nil, Err(errQuery)
	}

	defer rows.Close()

	for rows.Next() {
		var (
			banNet   BanCIDR
			sourceID int64
			targetID int64
		)

		if errScan := rows.
			Scan(&banNet.NetID, &banNet.CIDR, &banNet.Origin,
				&banNet.CreatedOn, &banNet.UpdatedOn, &banNet.Reason, &banNet.ReasonText,
				&banNet.ValidUntil, &banNet.Deleted, &banNet.Note, &banNet.UnbanReasonText,
				&banNet.IsEnabled, &targetID, &sourceID, &banNet.AppealState); errScan != nil {
			return nil, Err(errScan)
		}

		banNet.SourceID = steamid.New(sourceID)
		banNet.TargetID = steamid.New(targetID)

		nets = append(nets, banNet)
	}

	return nets, nil
}

func (db *Store) GetBanNetByID(ctx context.Context, netID int64, banNet *BanCIDR) error {
	const query = `
		SELECT net_id, cidr, origin, created_on, updated_on, reason, reason_text, valid_until, deleted, 
		       note, unban_reason_text, is_enabled, target_id, source_id, appeal_state
		FROM ban_net
		WHERE deleted = false AND net_id = $1`

	var (
		sourceID int64
		targetID int64
	)

	errQuery := db.
		QueryRow(ctx, query, netID).
		Scan(&banNet.NetID, &banNet.CIDR, &banNet.Origin,
			&banNet.CreatedOn, &banNet.UpdatedOn, &banNet.Reason, &banNet.ReasonText,
			&banNet.ValidUntil, &banNet.Deleted, &banNet.Note, &banNet.UnbanReasonText,
			&banNet.IsEnabled, &targetID, &sourceID, &banNet.AppealState)
	if errQuery != nil {
		return Err(errQuery)
	}

	banNet.SourceID = steamid.New(sourceID)
	banNet.TargetID = steamid.New(targetID)

	return nil
}

// GetBansNet returns the BanCIDR matching intersecting the supplied ip.
func (db *Store) GetBansNet(ctx context.Context) ([]BanCIDR, error) {
	const query = `
		SELECT net_id, cidr, origin, created_on, updated_on, reason, reason_text, valid_until, deleted, 
		       note, unban_reason_text, is_enabled, target_id, source_id, appeal_state 
		FROM ban_net
		WHERE deleted = false`

	var nets []BanCIDR

	rows, errQuery := db.Query(ctx, query)
	if errQuery != nil {
		return nil, Err(errQuery)
	}

	defer rows.Close()

	for rows.Next() {
		var (
			banNet   BanCIDR
			sourceID int64
			targetID int64
		)

		if errScan := rows.
			Scan(&banNet.NetID, &banNet.CIDR, &banNet.Origin,
				&banNet.CreatedOn, &banNet.UpdatedOn, &banNet.Reason, &banNet.ReasonText,
				&banNet.ValidUntil, &banNet.Deleted, &banNet.Note, &banNet.UnbanReasonText,
				&banNet.IsEnabled, &targetID, &sourceID, &banNet.AppealState); errScan != nil {
			return nil, Err(errScan)
		}

		banNet.SourceID = steamid.New(sourceID)
		banNet.TargetID = steamid.New(targetID)

		nets = append(nets, banNet)
	}

	return nets, nil
}

func (db *Store) updateBanNet(ctx context.Context, banNet *BanCIDR) error {
	banNet.UpdatedOn = time.Now()

	query, args, errQueryArgs := db.sb.
		Update("ban_net").
		Set("cidr", banNet.CIDR).
		Set("origin", banNet.Origin).
		Set("updated_on", banNet.UpdatedOn).
		Set("reason", banNet.Reason).
		Set("reason_text", banNet.ReasonText).
		Set("valid_until", banNet.ValidUntil).
		Set("deleted", banNet.Deleted).
		Set("note", banNet.Note).
		Set("unban_reason_text", banNet.UnbanReasonText).
		Set("is_enabled", banNet.IsEnabled).
		Set("target_id", banNet.TargetID.Int64()).
		Set("source_id", banNet.SourceID.Int64()).
		Set("appeal_state", banNet.AppealState).
		Where(sq.Eq{"net_id": banNet.NetID}).
		ToSql()
	if errQueryArgs != nil {
		return Err(errQueryArgs)
	}

	return Err(db.Exec(ctx, query, args...))
}

func (db *Store) insertBanNet(ctx context.Context, banNet *BanCIDR) error {
	query, args, errQueryArgs := db.sb.
		Insert("ban_net").
		Columns("cidr", "origin", "created_on", "updated_on", "reason", "reason_text", "valid_until",
			"deleted", "note", "unban_reason_text", "is_enabled", "target_id", "source_id", "appeal_state").
		Values(banNet.CIDR, banNet.Origin, banNet.CreatedOn, banNet.UpdatedOn, banNet.Reason, banNet.ReasonText,
			banNet.ValidUntil, banNet.Deleted, banNet.Note, banNet.UnbanReasonText, banNet.IsEnabled,
			banNet.TargetID.Int64(), banNet.SourceID.Int64(), banNet.AppealState).
		Suffix("RETURNING net_id").
		ToSql()
	if errQueryArgs != nil {
		return Err(errQueryArgs)
	}

	return Err(db.QueryRow(ctx, query, args...).Scan(&banNet.NetID))
}

func (db *Store) SaveBanNet(ctx context.Context, banNet *BanCIDR) error {
	if banNet.NetID > 0 {
		return db.updateBanNet(ctx, banNet)
	}

	return db.insertBanNet(ctx, banNet)
}

func (db *Store) DropBanNet(ctx context.Context, banNet *BanCIDR) error {
	query, args, errQueryArgs := db.sb.
		Delete("ban_net").
		Where(sq.Eq{"net_id": banNet.NetID}).
		ToSql()
	if errQueryArgs != nil {
		return Err(errQueryArgs)
	}

	if errExec := db.Exec(ctx, query, args...); errExec != nil {
		return Err(errExec)
	}

	banNet.NetID = 0

	return nil
}

func (db *Store) GetExpiredNetBans(ctx context.Context) ([]BanCIDR, error) {
	const query = `
		SELECT net_id, cidr, origin, created_on, updated_on, reason_text, valid_until, deleted, note, 
		       unban_reason_text, is_enabled, target_id, source_id, reason, appeal_state
		FROM ban_net
		WHERE valid_until < $1`

	var bans []BanCIDR

	rows, errQuery := db.Query(ctx, query, time.Now())
	if errQuery != nil {
		return nil, Err(errQuery)
	}

	defer rows.Close()

	for rows.Next() {
		var (
			banNet   BanCIDR
			targetID int64
			sourceID int64
		)

		if errScan := rows.
			Scan(&banNet.NetID, &banNet.CIDR, &banNet.Origin, &banNet.CreatedOn,
				&banNet.UpdatedOn, &banNet.ReasonText, &banNet.ValidUntil, &banNet.Deleted, &banNet.Note,
				&banNet.UnbanReasonText, &banNet.IsEnabled, &targetID, &sourceID,
				&banNet.Reason, &banNet.AppealState); errScan != nil {
			return nil, Err(errScan)
		}

		banNet.TargetID = steamid.New(targetID)
		banNet.SourceID = steamid.New(sourceID)

		bans = append(bans, banNet)
	}

	return bans, nil
}

func (db *Store) GetExpiredASNBans(ctx context.Context) ([]BanASN, error) {
	const query = `
		SELECT ban_asn_id, as_num, origin, source_id, target_id, reason_text, valid_until, created_on, updated_on, 
		       deleted, reason, is_enabled, unban_reason_text, appeal_state
		FROM ban_asn
		WHERE valid_until < $1 AND deleted = false`

	var bans []BanASN

	rows, errQuery := db.conn.Query(ctx, query, time.Now())
	if errQuery != nil {
		return nil, Err(errQuery)
	}

	defer rows.Close()

	for rows.Next() {
		var (
			banASN   BanASN
			targetID int64
			sourceID int64
		)

		if errScan := rows.
			Scan(&banASN.BanASNId, &banASN.ASNum, &banASN.Origin, &sourceID, &targetID,
				&banASN.ReasonText, &banASN.ValidUntil, &banASN.CreatedOn, &banASN.UpdatedOn, &banASN.Deleted,
				&banASN.Reason, &banASN.IsEnabled, &banASN.UnbanReasonText, &banASN.AppealState); errScan != nil {
			return nil, errors.Wrap(errScan, "Failed to scan asn ban")
		}

		banASN.TargetID = steamid.New(targetID)
		banASN.SourceID = steamid.New(sourceID)

		bans = append(bans, banASN)
	}

	return bans, nil
}

func (db *Store) GetASNRecordsByNum(ctx context.Context, asNum int64) (ip2location.ASNRecords, error) {
	const query = `
		SELECT ip_from, ip_to, cidr, as_num, as_name 
		FROM net_asn
		WHERE as_num = $1`

	rows, errQuery := db.conn.Query(ctx, query, asNum)
	if errQuery != nil {
		return nil, Err(errQuery)
	}

	defer rows.Close()

	var records ip2location.ASNRecords

	for rows.Next() {
		var asnRecord ip2location.ASNRecord
		if errScan := rows.
			Scan(&asnRecord.IPFrom, &asnRecord.IPTo, &asnRecord.CIDR, &asnRecord.ASNum, &asnRecord.ASName); errScan != nil {
			return nil, Err(errScan)
		}

		records = append(records, asnRecord)
	}

	return records, nil
}

func (db *Store) GetASNRecordByIP(ctx context.Context, ipAddr net.IP, asnRecord *ip2location.ASNRecord) error {
	const query = `
		SELECT ip_from, ip_to, cidr, as_num, as_name 
		FROM net_asn
		WHERE $1::inet <@ ip_range
		LIMIT 1`

	if errQuery := db.conn.
		QueryRow(ctx, query, ipAddr.String()).
		Scan(&asnRecord.IPFrom, &asnRecord.IPTo, &asnRecord.CIDR, &asnRecord.ASNum, &asnRecord.ASName); errQuery != nil {
		return Err(errQuery)
	}

	return nil
}

func (db *Store) GetLocationRecord(ctx context.Context, ipAddr net.IP, record *ip2location.LocationRecord) error {
	const query = `
		SELECT ip_from, ip_to, country_code, country_name, region_name, city_name, ST_Y(location), ST_X(location) 
		FROM net_location 
		WHERE ip_range @> $1::inet`

	if errQuery := db.QueryRow(ctx, query, ipAddr.String()).
		Scan(&record.IPFrom, &record.IPTo, &record.CountryCode, &record.CountryName, &record.RegionName,
			&record.CityName, &record.LatLong.Latitude, &record.LatLong.Longitude); errQuery != nil {
		return Err(errQuery)
	}

	return nil
}

func (db *Store) GetProxyRecord(ctx context.Context, ipAddr net.IP, proxyRecord *ip2location.ProxyRecord) error {
	const query = `
		SELECT ip_from, ip_to, proxy_type, country_code, country_name, region_name, 
       		city_name, isp, domain_used, usage_type, as_num, as_name, last_seen, threat 
		FROM net_proxy 
		WHERE $1::inet <@ ip_range`

	if errQuery := db.QueryRow(ctx, query, ipAddr.String()).
		Scan(&proxyRecord.IPFrom, &proxyRecord.IPTo, &proxyRecord.ProxyType, &proxyRecord.CountryCode, &proxyRecord.CountryName, &proxyRecord.RegionName, &proxyRecord.CityName, &proxyRecord.ISP,
			&proxyRecord.Domain, &proxyRecord.UsageType, &proxyRecord.ASN, &proxyRecord.AS, &proxyRecord.LastSeen, &proxyRecord.Threat); errQuery != nil {
		return Err(errQuery)
	}

	return nil
}

func (db *Store) loadASN(ctx context.Context, records []ip2location.ASNRecord) error {
	curTime := time.Now()

	if errTruncate := db.truncateTable(ctx, tableNetASN); errTruncate != nil {
		return errTruncate
	}

	const query = `
		INSERT INTO net_asn (ip_from, ip_to, cidr, as_num, as_name, ip_range) 
		VALUES($1, $2, $3, $4, $5, iprange($1, $2))`

	batch := pgx.Batch{}

	for recordIdx, asnRecord := range records {
		batch.Queue(query, asnRecord.IPFrom, asnRecord.IPTo, asnRecord.CIDR, asnRecord.ASNum, asnRecord.ASName)

		if recordIdx > 0 && recordIdx%100000 == 0 || len(records) == recordIdx+1 {
			if batch.Len() > 0 {
				c, cancel := context.WithTimeout(ctx, time.Second*10)

				batchResults := db.conn.SendBatch(c, &batch)
				if errCloseBatch := batchResults.Close(); errCloseBatch != nil {
					cancel()

					return errors.Wrapf(errCloseBatch, "Failed to close asn batch")
				}

				cancel()

				batch = pgx.Batch{}

				db.log.Info(fmt.Sprintf("ASN Progress: %d/%d (%.0f%%)",
					recordIdx, len(records)-1, float64(recordIdx)/float64(len(records)-1)*100))
			}
		}
	}

	db.log.Info("Loaded ASN4 records",
		zap.Int("count", len(records)), zap.Duration("duration", time.Since(curTime)))

	return nil
}

func (db *Store) loadLocation(ctx context.Context, records []ip2location.LocationRecord, _ bool) error {
	curTime := time.Now()

	if errTruncate := db.truncateTable(ctx, tableNetLocation); errTruncate != nil {
		return errTruncate
	}

	const query = `
		INSERT INTO net_location (ip_from, ip_to, country_code, country_name, region_name, city_name, location, ip_range)
		VALUES($1, $2, $3, $4, $5, $6, ST_SetSRID(ST_MakePoint($8, $7), 4326), iprange($1, $2))`

	batch := pgx.Batch{}

	for recordIdx, locationRecord := range records {
		batch.Queue(query, locationRecord.IPFrom, locationRecord.IPTo, locationRecord.CountryCode, locationRecord.CountryName, locationRecord.RegionName, locationRecord.CityName, locationRecord.LatLong.Latitude, locationRecord.LatLong.Longitude)

		if recordIdx > 0 && recordIdx%100000 == 0 || len(records) == recordIdx+1 {
			if batch.Len() > 0 {
				c, cancel := context.WithTimeout(ctx, time.Second*10)

				batchResults := db.conn.SendBatch(c, &batch)
				if errCloseBatch := batchResults.Close(); errCloseBatch != nil {
					cancel()

					return errors.Wrapf(errCloseBatch, "Failed to send location batch update query")
				}

				cancel()

				batch = pgx.Batch{}

				db.log.Info(fmt.Sprintf("Location4 Progress: %d/%d (%.0f%%)",
					recordIdx, len(records)-1, float64(recordIdx)/float64(len(records)-1)*100))
			}
		}
	}

	db.log.Info("Loaded Location4 records",
		zap.Int("count", len(records)), zap.Duration("duration", time.Since(curTime)))

	return nil
}

func (db *Store) loadProxies(ctx context.Context, records []ip2location.ProxyRecord, _ bool) error {
	curTime := time.Now()

	if errTruncate := db.truncateTable(ctx, tableNetProxy); errTruncate != nil {
		return errTruncate
	}

	const query = `
		INSERT INTO net_proxy (ip_from, ip_to, proxy_type, country_code, country_name, region_name, city_name, isp,
		                       domain_used, usage_type, as_num, as_name, last_seen, threat, ip_range)
		VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, iprange($1, $2))`

	batch := pgx.Batch{}

	for recordIdx, proxyRecord := range records {
		batch.Queue(query, proxyRecord.IPFrom, proxyRecord.IPTo, proxyRecord.ProxyType, proxyRecord.CountryCode, proxyRecord.CountryName, proxyRecord.RegionName, proxyRecord.CityName,
			proxyRecord.ISP, proxyRecord.Domain, proxyRecord.UsageType, proxyRecord.ASN, proxyRecord.AS, proxyRecord.LastSeen, proxyRecord.Threat)

		if recordIdx > 0 && recordIdx%100000 == 0 || len(records) == recordIdx+1 {
			if batch.Len() > 0 {
				c, cancel := context.WithTimeout(ctx, time.Second*120)

				batchResults := db.conn.SendBatch(c, &batch)
				if errCloseBatch := batchResults.Close(); errCloseBatch != nil {
					cancel()

					return errors.Wrapf(errCloseBatch, "Faield to close proxy batch")
				}

				cancel()

				batch = pgx.Batch{}

				db.log.Info(fmt.Sprintf("Proxy Progress: %d/%d (%.0f%%)",
					recordIdx, len(records)-1, float64(recordIdx)/float64(len(records)-1)*100))
			}
		}
	}

	db.log.Info("Loaded Proxy records",
		zap.Int("count", len(records)), zap.Duration("duration", time.Since(curTime)))

	return nil
}

// InsertBlockListData will load the provided datasets into the database
//
// Note that this can take a while on slower machines. For reference, it takes
// about ~90s with a local database on a Ryzen 3900X/PCIe4 NVMe SSD.
func (db *Store) InsertBlockListData(ctx context.Context, blockListData *ip2location.BlockListData) error {
	if len(blockListData.Proxies) > 0 {
		if errProxies := db.loadProxies(ctx, blockListData.Proxies, false); errProxies != nil {
			return errProxies
		}
	}

	if len(blockListData.Locations4) > 0 {
		if errLocation := db.loadLocation(ctx, blockListData.Locations4, false); errLocation != nil {
			return errLocation
		}
	}

	if len(blockListData.ASN4) > 0 {
		if errASN := db.loadASN(ctx, blockListData.ASN4); errASN != nil {
			return errASN
		}
	}

	return nil
}

func (db *Store) GetBanASN(ctx context.Context, asNum int64, banASN *BanASN) error {
	const query = `
		SELECT ban_asn_id, as_num, origin, source_id, target_id, reason_text, valid_until, created_on, updated_on, 
		       deleted, reason, is_enabled, unban_reason_text, appeal_state
		FROM ban_asn 
		WHERE deleted = false AND as_num = $1`

	var (
		targetID int64
		sourceID int64
	)

	if errQuery := db.
		QueryRow(ctx, query, asNum).
		Scan(&banASN.BanASNId, &banASN.ASNum, &banASN.Origin,
			&sourceID, &targetID, &banASN.ReasonText, &banASN.ValidUntil, &banASN.CreatedOn,
			&banASN.UpdatedOn, &banASN.Deleted, &banASN.Reason, &banASN.IsEnabled, &banASN.UnbanReasonText,
			&banASN.AppealState); errQuery != nil {
		return Err(errQuery)
	}

	banASN.TargetID = steamid.New(targetID)
	banASN.SourceID = steamid.New(sourceID)

	return nil
}

func (db *Store) GetBansASN(ctx context.Context) ([]BanASN, error) {
	const query = `
		SELECT ban_asn_id, as_num, origin, source_id, target_id, reason_text, valid_until, created_on, updated_on, 
		       deleted, reason, is_enabled, unban_reason_text, appeal_state
		FROM ban_asn 
		WHERE deleted = false`

	rows, errRows := db.Query(ctx, query)
	if errRows != nil {
		return nil, Err(errRows)
	}

	defer rows.Close()

	var records []BanASN

	for rows.Next() {
		var (
			ban      BanASN
			targetID int64
			sourceID int64
		)

		if errQuery := rows.
			Scan(&ban.BanASNId, &ban.ASNum, &ban.Origin, &sourceID, &targetID, &ban.ReasonText, &ban.ValidUntil,
				&ban.CreatedOn, &ban.UpdatedOn, &ban.Deleted, &ban.Reason, &ban.IsEnabled, &ban.UnbanReasonText, &ban.AppealState); errQuery != nil {
			return nil, Err(errQuery)
		}

		ban.SourceID = steamid.New(sourceID)
		ban.TargetID = steamid.New(targetID)

		records = append(records, ban)
	}

	return records, nil
}

func (db *Store) SaveBanASN(ctx context.Context, banASN *BanASN) error {
	banASN.UpdatedOn = time.Now()

	if banASN.BanASNId > 0 {
		const queryUpdate = `
			UPDATE ban_asn 
			SET as_num = $2, origin = $3, source_id = $4, target_id = $5, reason = $6,
				valid_until = $7, updated_on = $8, reason_text = $9, is_enabled = $10, deleted = $11, 
				unban_reason_text = $12, appeal_state = $13
			WHERE ban_asn_id = $1`

		return Err(db.
			Exec(ctx, queryUpdate, banASN.BanASNId, banASN.ASNum, banASN.Origin, banASN.SourceID.Int64(),
				banASN.TargetID.Int64(), banASN.Reason, banASN.ValidUntil, banASN.UpdatedOn, banASN.ReasonText, banASN.IsEnabled,
				banASN.Deleted, banASN.UnbanReasonText, banASN.AppealState))
	}

	const queryInsert = `
		INSERT INTO ban_asn (as_num, origin, source_id, target_id, reason, valid_until, updated_on, created_on, 
		                     reason_text, is_enabled, deleted, unban_reason_text, appeal_state)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING ban_asn_id`

	return Err(db.
		QueryRow(ctx, queryInsert, banASN.ASNum, banASN.Origin, banASN.SourceID.Int64(), banASN.TargetID.Int64(),
			banASN.Reason, banASN.ValidUntil, banASN.UpdatedOn, banASN.CreatedOn, banASN.ReasonText, banASN.IsEnabled,
			banASN.Deleted, banASN.UnbanReasonText, banASN.AppealState).
		Scan(&banASN.BanASNId))
}

func (db *Store) DropBanASN(ctx context.Context, banASN *BanASN) error {
	banASN.Deleted = true

	return db.SaveBanASN(ctx, banASN)
}

func (db *Store) GetSteamIDsAtIP(ctx context.Context, ipNet *net.IPNet) (steamid.Collection, error) {
	const query = `
		SELECT DISTINCT c.steam_id
		FROM person_connections c
		WHERE ip_addr::inet <<= inet '%s';`

	if ipNet == nil {
		return nil, errors.New("Invalid address")
	}

	rows, errQuery := db.Query(ctx, fmt.Sprintf(query, ipNet.String()))
	if errQuery != nil {
		return nil, Err(errQuery)
	}

	defer rows.Close()

	var ids steamid.Collection

	for rows.Next() {
		var sid64 int64
		if errScan := rows.Scan(&sid64); errScan != nil {
			return nil, Err(errScan)
		}

		ids = append(ids, steamid.New(sid64))
	}

	return ids, nil
}
