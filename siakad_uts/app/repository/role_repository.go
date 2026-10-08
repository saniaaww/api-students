package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RoleRepository interface {
	LoadPermissions(ctx context.Context) (map[string][]string, error)
}

type roleRepository struct {
	db *pgxpool.Pool
}

func NewRoleRepository(db *pgxpool.Pool) RoleRepository {
	return &roleRepository{
		db: db,
	}
}

func (r *roleRepository) LoadPermissions(
	ctx context.Context,
) (map[string][]string, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			r.name AS role_name,
			p.name AS permission_name
		FROM roles r
		LEFT JOIN role_permissions rp
			ON rp.role_name = r.name
		LEFT JOIN permissions p
			ON p.name = rp.permission_name
		ORDER BY r.name, p.name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string][]string)

	for rows.Next() {
		var (
			roleName       string
			permissionName *string
		)

		if err := rows.Scan(&roleName, &permissionName); err != nil {
			return nil, err
		}

		// Pastikan setiap role tetap masuk
		// meskipun belum memiliki permission.
		if _, ok := result[roleName]; !ok {
			result[roleName] = []string{}
		}

		if permissionName != nil {
			result[roleName] = append(
				result[roleName],
				*permissionName,
			)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
