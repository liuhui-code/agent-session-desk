package store

import (
	"context"
	"database/sql"
)

func validateGroupScopes(ctx context.Context, tx *sql.Tx) error {
	rows, err := tx.QueryContext(ctx, `SELECT scope FROM sessions UNION SELECT scope FROM groups UNION SELECT scope FROM session_groups`)
	if err != nil {
		return err
	}
	var scopes []string
	for rows.Next() {
		var scope string
		if err = rows.Scan(&scope); err != nil {
			rows.Close()
			return err
		}
		scopes = append(scopes, scope)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	for _, scope := range scopes {
		if _, err = (&Store{scope: scope}).groups(ctx, tx); err != nil {
			return err
		}
	}
	return nil
}
