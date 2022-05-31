package storage

const insertStmt = `WITH e AS (
			INSERT INTO short_url (original_url, short_url, user_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (original_url) DO NOTHING
		RETURNING short_url
		)
		SELECT short_url, 100000
		FROM e
		UNION ALL
		SELECT short_url, 100001
		FROM short_url
		WHERE original_url=$1`
