package db

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"sort"
)

//go:embed migrations/*.sql
var migrations embed.FS

func Migrate(ctx context.Context, d *DB) error {
	entries, err := fs.ReadDir(migrations, "migrations")
	if err != nil {
		return err
	}

	// 按文件名排序，确保 0001、0002…顺序执行
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		b, err := migrations.ReadFile("migrations/" + e.Name())
		if err != nil {
			return err
		}
		if _, err := d.Pool.Exec(ctx, string(b)); err != nil {
			return fmt.Errorf("migration %s: %w", e.Name(), err)
		}
	}
	return nil
}
