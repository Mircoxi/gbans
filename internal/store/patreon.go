package store

import (
	"context"

	"github.com/pkg/errors"
)

func (db *Store) SetPatreonAuth(ctx context.Context, accessToken string, refreshToken string) error {
	query, args, errQuery := db.sb.
		Update("patreon_auth").
		Set("creator_access_token", accessToken).
		Set("creator_refresh_token", refreshToken).ToSql()
	if errQuery != nil {
		return errors.Wrap(errQuery, "Failed to create patreon auth update query")
	}

	return Err(db.Exec(ctx, query, args...))
}

func (db *Store) GetPatreonAuth(ctx context.Context) (string, string, error) {
	query, args, errQuery := db.sb.
		Select("creator_access_token", "creator_refresh_token").From("patreon_auth").ToSql()
	if errQuery != nil {
		return "", "", errors.Wrap(errQuery, "Failed to create patreon auth select query")
	}

	var (
		creatorAccessToken  string
		creatorRefreshToken string
	)

	if errScan := db.QueryRow(ctx, query, args...).Scan(&creatorAccessToken, &creatorRefreshToken); errScan != nil {
		return "", "", errors.Wrap(errQuery, "Failed to query patreon auth")
	}

	return creatorAccessToken, creatorRefreshToken, nil
}
