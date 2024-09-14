package test

import (
	"context"
	"goutils/sqldb"
	"goutils/sqldb/test/models"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

func testUpdate(t *testing.T, db *sqldb.DB) {
	r := require.New(t)
	updateData := map[string]any{
		"age": 55,
	}
	res, err := db.Table("users").Where("name", "=", "foo").Update(updateData)
	r.NoError(err)
	affected, err := res.RowsAffected()
	r.NoError(err)
	r.Equal(int64(1), affected)

	rows, err := db.Query("SELECT * FROM users where name = ?", "foo")
	r.NoError(err)
	defer rows.Close()
	r.True(rows.Next())
	var name string
	var age int
	err = rows.Scan(&name, &age)
	r.NoError(err)
	r.Equal("foo", name)
	r.Equal(55, age)
}

func testSelect(t *testing.T, db *sqldb.DB) {
	r := require.New(t)
	var models []models.Users
	err := db.StructScan(&models, "SELECT * FROM users")
	r.NoError(err)
	r.Len(models, 1)
	r.Equal("foo", models[0].Name.String)
	r.Equal(int64(55), models[0].Age.Int64)

	//TODO fix this
	var ages []struct {
		Age int `db:"age"`
	}
	err = db.StructScan(&ages, "SELECT age FROM users")
	r.NoError(err)
	r.Len(ages, 1)
	r.Equal(55, ages[0].Age)
}

func testQuery(t *testing.T, db *sqldb.DB) {
	r := require.New(t)
	rows, err := db.Query("SELECT * FROM users where name = ?", "foo")
	r.NoError(err)
	defer rows.Close()
	r.True(rows.Next())
	var name string
	var age int
	err = rows.Scan(&name, &age)
	r.NoError(err)
	r.Equal("foo", name)
	r.Equal(20, age)
}

func testQueryContext(t *testing.T, db *sqldb.DB, ctx context.Context) {
	r := require.New(t)
	rows, err := db.QueryContext(ctx, "SELECT * FROM users where name = ?", "foo")
	r.NoError(err)
	defer rows.Close()
	r.True(rows.Next())
	var name string
	var age int
	err = rows.Scan(&name, &age)
	r.NoError(err)
	r.Equal("foo", name)
	r.Equal(20, age)
}

func testQueryRow(t *testing.T, db *sqldb.DB) {
	r := require.New(t)
	row := db.QueryRow("SELECT * FROM users where name = ?", "foo")
	var name string
	var age int
	err := row.Scan(&name, &age)
	r.NoError(err)
	r.Equal("foo", name)
	r.Equal(20, age)
}
func testQueryRowContext(t *testing.T, db *sqldb.DB, ctx context.Context) {
	r := require.New(t)
	row := db.QueryRowContext(ctx, "SELECT * FROM users where name = ?", "foo")
	var name string
	var age int
	err := row.Scan(&name, &age)
	r.NoError(err)
	r.Equal("foo", name)
	r.Equal(20, age)
}

func testExec(t *testing.T, db *sqldb.DB) {
	r := require.New(t)
	res, err := db.Exec("INSERT INTO users (name, age) VALUES (?, ?)", "luci", 25)
	r.NoError(err)
	affected, err := res.RowsAffected()
	r.NoError(err)
	r.Equal(int64(1), affected)
}

func testExecContext(t *testing.T, db *sqldb.DB, ctx context.Context) {
	r := require.New(t)
	res, err := db.ExecContext(ctx, "DELETE FROM users where name = ?", "luci")
	r.NoError(err)
	affected, err := res.RowsAffected()
	r.NoError(err)
	r.Equal(int64(1), affected)
}

func testGet(t *testing.T, db *sqldb.DB) {
	r := require.New(t)
	var age int
	err := sqldb.Get(db, &age, "SELECT age FROM users where name = ?", "foo")
	r.NoError(err)
	r.Equal(55, age)
	var model models.Users
	err = db.Get(&model, "SELECT * FROM users where name = ?", "foo")
	r.NoError(err)
	r.Equal("foo", model.Name.String)
	r.Equal(int64(55), model.Age.Int64)
	err = db.Table("users").Where("name", "=", "foo").ScanRow(&model)
	r.NoError(err)
	r.Equal("foo", model.Name.String)
	r.Equal(int64(55), model.Age.Int64)
}

func TestDB(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()
	opts := []sqldb.Option{
		sqldb.WithTraceSQL(true),
		sqldb.WithDebug(true),
	}
	t.Run("sqlite3", func(t *testing.T) {
		db, err := sqldb.Open("sqlite3", ":memory:", opts...)
		r.NoError(err)
		_, err = db.Exec("CREATE TABLE users (name TEXT, age INTEGER)")
		r.NoError(err)
		res, err := db.Table("users").Insert(map[string]any{
			"name": "foo",
			"age":  20,
		})
		r.NoError(err)
		r.NotNil(res)
		affected, err := res.RowsAffected()
		r.NoError(err)
		r.Equal(int64(1), affected)
		testQuery(t, db)
		testQueryContext(t, db, ctx)
		testQueryRow(t, db)
		testQueryRowContext(t, db, ctx)
		testExec(t, db)
		testExecContext(t, db, ctx)
		testUpdate(t, db)
		testSelect(t, db)
		testGet(t, db)
	})
	t.Run("mysql", func(t *testing.T) {
		db, err := sqldb.Open("mysql", "root:admin@tcp(127.0.0.1:3306)/test")
		r.NoError(err)
		_, err = db.Exec("DROP TABLE IF EXISTS users")
		r.NoError(err)
		_, err = db.Exec("CREATE TABLE users (name TEXT, age INTEGER)")
		r.NoError(err)
		res, err := db.Table("users").Insert(map[string]any{
			"name": "foo",
			"age":  20,
		})
		r.NoError(err)
		r.NotNil(res)
		affected, err := res.RowsAffected()
		r.NoError(err)
		r.Equal(int64(1), affected)

		testQuery(t, db)
		testQueryContext(t, db, ctx)
		testQueryRow(t, db)
		testQueryRowContext(t, db, ctx)
		testExec(t, db)
		testExecContext(t, db, ctx)
		testUpdate(t, db)
		testSelect(t, db)
		testGet(t, db)
	})
	t.Run("postgres", func(t *testing.T) {
		db, err := sqldb.Open("postgres", "user=postgres password=admin dbname=postgres sslmode=disable")
		r.NoError(err)
		_, err = db.Exec("DROP TABLE IF EXISTS users")
		r.NoError(err)
		_, err = db.Exec("CREATE TABLE users (name TEXT, age INTEGER)")
		r.NoError(err)
		res, err := db.Table("users").Insert(map[string]any{
			"name": "foo",
			"age":  20,
		})
		r.NoError(err)
		r.NotNil(res)
		affected, err := res.RowsAffected()
		r.NoError(err)
		r.Equal(int64(1), affected)

		testQuery(t, db)
		testQueryContext(t, db, ctx)
		testQueryRow(t, db)
		testQueryRowContext(t, db, ctx)
		testExec(t, db)
		testExecContext(t, db, ctx)
		testUpdate(t, db)
		testSelect(t, db)
		testGet(t, db)
	})
}

func BenchmarkInsert(b *testing.B) {
	exec := func(db *sqldb.DB, b *testing.B) {
		_, err := db.Exec("DROP TABLE IF EXISTS users")
		if err != nil {
			b.Fatal(err)
		}
		_, err = db.Exec("CREATE TABLE users (name TEXT, age INTEGER)")
		if err != nil {
			b.Fatal(err)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err = db.Table("users").Insert(map[string]any{
				"name": "foo",
				"age":  20,
			})
			if err != nil {
				b.Fatal(err)
			}
		}
	}
	b.Run("sqlite3", func(b *testing.B) {
		db, err := sqldb.Open("sqlite3", ":memory:")
		if err != nil {
			b.Fatal(err)
		}
		exec(db, b)
	})
	b.Run("postgres", func(b *testing.B) {
		db, err := sqldb.Open("postgres", "user=postgres password=admin dbname=postgres sslmode=disable")
		if err != nil {
			b.Fatal(err)
		}
		exec(db, b)
	})
	b.Run("mysql", func(b *testing.B) {
		db, err := sqldb.Open("mysql", "root:admin@tcp(127.0.0.1:3306)/test")
		if err != nil {
			b.Fatal(err)
		}
		exec(db, b)
	})
}
