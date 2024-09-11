package sqldb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

var stringBuilderPool = sync.Pool{
	New: func() any {
		return new(strings.Builder)
	}}

func acquireStringBuilder() *strings.Builder {
	return stringBuilderPool.Get().(*strings.Builder)
}

func releaseStringBuilder(b *strings.Builder) {
	b.Reset()
	stringBuilderPool.Put(b)
}

type DatabaseProvider interface {
	OptionProvider
	Execer
	Queryer
}

type Execer interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

type Queryer interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type OptionProvider interface {
	Option() option
}

func Insert(ctx context.Context, flavor Flavor, prefix string, execer Execer, table string, data map[string]any) (sql.Result, error) {
	dataLen := len(data)
	if dataLen == 0 {
		return nil, fmt.Errorf("no data to insert")
	}
	fields := make([]string, 0, dataLen)
	values := make([]any, 0, len(data))
	for k, v := range data {
		fields = append(fields, flavor.columnQuote(k))
		values = append(values, v)
	}
	placeholder := flavor.placeHolder(dataLen)

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", flavor.tableQuote(prefix, table), strings.Join(fields, ","), placeholder)
	return execer.ExecContext(ctx, query, values...)
}

func Select(ctx context.Context, queryer Queryer, flavor Flavor, prefix, table string, columns string, where Conditions) (*sql.Rows, error) {
	var (
		whereClause string
		whereArgs   []any
	)
	conditionsCalsue(flavor, where, &whereClause, &whereArgs)
	if whereClause != "" {
		whereClause = " WHERE " + whereClause
	}
	query := fmt.Sprintf("SELECT %s FROM %s%s", columns, flavor.tableQuote(prefix, table), whereClause)
	return queryer.QueryContext(ctx, query, whereArgs...)
}

type aggregateType string

const (
	aggregateTypeCount aggregateType = "COUNT"
	aggregateTypeSum   aggregateType = "SUM"
	aggregateTypeAvg   aggregateType = "AVG"
	aggregateTypeMax   aggregateType = "MAX"
	aggregateTypeMin   aggregateType = "MIN"
)

func aggregate[T DatabaseProvider](ctx context.Context, t T, at aggregateType, table, column, where string, args []any) (int, error) {
	// where, _, err := conditionWhere(t, where, args)
	// if err != nil {
	// 	return 0, err
	// }
	// query := fmt.Sprintf("SELECT %s(%s) FROM %s WHERE %s", at, columnQuote(t, column), tableQuote(t, table), where)
	// var count int

	// err = t.QueryRowContext(ctx, query, args...).Scan(&count)
	// return count, err
	return 0, nil
}

func fixQuery(flavor Flavor, query string) string {
	builder := acquireStringBuilder()
	switch flavor {
	case MySQL, SQLite:
		return query
	}
	defer releaseStringBuilder(builder)
	var i, j int
	for i = strings.IndexRune(query, '?'); i != -1; i = strings.IndexRune(query, '?') {
		j++
		builder.WriteString(query[:i])
		switch flavor {
		case PostgreSQL:
			builder.WriteString("$" + strconv.Itoa(j))
		}
		query = query[i+1:]
	}
	builder.WriteString(query)
	return builder.String()
}

func Count(ctx context.Context, queryer Queryer, flavor Flavor, prefix, table string, where string, args ...any) (int, error) {
	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", flavor.tableQuote(prefix, table), where)
	err := queryer.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

func Update(ctx context.Context, flavor Flavor, prefix string, execer Execer, table string, data map[string]any, where Conditions) (sql.Result, error) {
	if len(where) == 0 {
		return nil, fmt.Errorf("where conditions is empty")
	}
	dataLen := len(data)
	if dataLen == 0 {
		return nil, fmt.Errorf("no data to update")
	}
	fields := make([]string, 0, dataLen)
	values := make([]any, 0, dataLen)
	for k, v := range data {
		fields = append(fields, fmt.Sprintf("%s=?", flavor.columnQuote(k)))
		values = append(values, v)
	}
	var (
		whereClause string
		whereArgs   []any
	)
	conditionsCalsue(flavor, where, &whereClause, &whereArgs)
	if whereClause != "" {
		whereClause = " WHERE " + whereClause
	}

	query := fmt.Sprintf("UPDATE %s SET %s%s", flavor.tableQuote(prefix, table), strings.Join(fields, ", "), whereClause)
	values = append(values, whereArgs...)
	return execer.ExecContext(ctx, query, values...)
}

func FormatSQL(query string, args []any) string {
	builder := acquireStringBuilder()
	defer releaseStringBuilder(builder)
	nArgs := len(args)
	if nArgs == 0 {
		return query
	}
	var i, j int
	for i = strings.IndexRune(query, '?'); i != -1; i = strings.IndexRune(query, '?') {
		builder.WriteString(query[:i])
		switch a := args[j].(type) {
		case *int64:
			val := args[i]
			if val.(*int64) != nil {
				builder.WriteString(fmt.Sprintf("%d", *val.(*int64)))
			} else {
				builder.WriteString("NULL")
			}
		case *int:
			val := args[i]
			if val.(*int) != nil {
				builder.WriteString(fmt.Sprintf("%d", *val.(*int)))
			} else {
				builder.WriteString("NULL")
			}
		case *float64:
			val := args[i]
			if val.(*float64) != nil {
				builder.WriteString(fmt.Sprintf("%f", *val.(*float64)))
			} else {
				builder.WriteString("NULL")
			}
		case *bool:
			val := args[i]
			if val.(*bool) != nil {
				builder.WriteString(fmt.Sprintf("%t", *val.(*bool)))
			} else {
				builder.WriteString("NULL")
			}
		case *string:
			val := args[i]
			if val.(*string) != nil {
				builder.WriteString(fmt.Sprintf("'%q'", *val.(*string)))
			} else {
				builder.WriteString("NULL")
			}
		case *time.Time:
			val := args[i]
			if val.(*time.Time) != nil {
				time := *val.(*time.Time)
				builder.WriteString(fmt.Sprintf("'%v'", time.Format("2006-01-02 15:04:05")))
			} else {
				builder.WriteString("NULL")
			}
		case int, int8, int16, int32, int64,
			uint, uint8, uint16, uint32, uint64:
			builder.WriteString(fmt.Sprintf("%d", a))
		case float64:
			builder.WriteString(fmt.Sprintf("%f", a))
		case bool:
			builder.WriteString(fmt.Sprintf("%t", a))
		case time.Time:
			builder.WriteString(fmt.Sprintf("'%v'", a.Format("2006-01-02 15:04:05")))
		case sql.NullBool:
			if a.Valid {
				builder.WriteString(fmt.Sprintf("%t", a.Bool))
			} else {
				builder.WriteString("NULL")
			}
		case sql.NullInt64:
			if a.Valid {
				builder.WriteString(fmt.Sprintf("%d", a.Int64))
			} else {
				builder.WriteString("NULL")
			}
		case sql.NullString:
			if a.Valid {
				builder.WriteString(fmt.Sprintf("%q", a.String))
			} else {
				builder.WriteString("NULL")
			}

		case nil:
			builder.WriteString("NULL")
		default:
			builder.WriteString(fmt.Sprintf("'%v'", a))
		}
		query = query[i+1:]
		j++
	}
	builder.WriteString(query)
	return builder.String()
}

func ParseSQLRow[T any](row *sql.Row) (T, error) {
	var schema T

	newSchema := reflect.New(reflect.TypeOf(schema)).Interface()

	s := reflect.ValueOf(newSchema).Elem()

	var fields []interface{}
	for i := 0; i < s.NumField(); i++ {
		fields = append(fields, s.Field(i).Addr().Interface())
	}

	err := row.Scan(fields...)
	if err != nil {
		return schema, err
	}

	return newSchema.(T), nil
}

// ParseSQLRows 解析多行数据并返回模型值切片
func ParseSQLRows[T any](rows *sql.Rows) ([]T, error) {
	var parsedRows []T

	_, _ = rows.Columns()
	// Fetch rows
	for rows.Next() {
		var schema T

		// 创建一个新的 T 类型的实例
		newSchema := reflect.New(reflect.TypeOf(schema)).Interface()

		// 获取新实例的反射值
		s := reflect.ValueOf(newSchema).Elem()

		// 创建一个字段地址的切片
		var fields []interface{}
		for i := 0; i < s.NumField(); i++ {
			fields = append(fields, s.Field(i).Addr().Interface())
		}

		// 扫描数据库行的值到字段地址中
		err := rows.Scan(fields...)
		if err != nil {
			return nil, err
		}

		// 将解析后的 T 类型实例添加到切片中
		parsedRows = append(parsedRows, newSchema.(T))
	}

	return parsedRows, nil
}

// ParseSQLRows 解析多行数据并返回模型值切片
func ParseSQLRows2[T any](rows *sql.Rows) ([]T, error) {
	var parsedRows []T

	// 获取列名
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	// 创建一个切片来存储列值
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))

	// Fetch rows
	for rows.Next() {
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		// 扫描数据库行的值到列值切片中
		err := rows.Scan(valuePtrs...)
		if err != nil {
			return nil, err
		}

		// 创建一个新的 T 类型的实例
		var schema T

		// 将列值赋值给 T 类型的实例
		err = assignValues(&schema, columns, values)
		if err != nil {
			return nil, err
		}

		// 将解析后的 T 类型实例添加到切片中
		parsedRows = append(parsedRows, schema)
	}

	return parsedRows, nil
}

// assignValues 将列值赋值给 T 类型的实例
func assignValues[T any](schema *T, columns []string, values []interface{}) error {
	s := reflect.ValueOf(schema).Elem()
	if s.Kind() != reflect.Struct {
		return errors.New("schema must be a struct")
	}

	for i, col := range columns {
		field := s.FieldByNameFunc(func(name string) bool {
			return col == name
		})
		if field.IsValid() && field.CanSet() {
			val := reflect.ValueOf(values[i])
			if val.Type().ConvertibleTo(field.Type()) {
				field.Set(val.Convert(field.Type()))
			} else {
				return errors.New("type mismatch for field " + col)
			}
		}
	}
	return nil
}

func scanRow(row *sql.Row, schema interface{}) (interface{}, error) {

	newSchema := reflect.New(reflect.ValueOf(schema).Elem().Type()).Interface()

	s := reflect.ValueOf(newSchema).Elem()

	var fields []interface{}
	for i := 0; i < s.NumField(); i++ {
		fields = append(fields, s.Field(i).Addr().Interface())
	}

	err := row.Scan(fields...)
	if err != nil {
		return nil, err
	}
	reflect.ValueOf(schema).Elem().Set(reflect.ValueOf(newSchema).Elem())
	return newSchema, nil
}

func Get[T any](q Queryer, dest T, query string, args ...interface{}) error {
	rows, err := q.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		return sql.ErrNoRows
	}

	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr || destValue.IsNil() {
		return errors.New("dest must be a non-nil pointer")
	}

	destElem := destValue.Elem()
	if destElem.Kind() == reflect.Struct {
		// 获取查询结果的列名
		columns, err := rows.Columns()
		if err != nil {
			return err
		}

		// 创建一个 map 来存储列名到结构体字段的映射
		fieldMap := make(map[string]interface{})
		for i := 0; i < destElem.NumField(); i++ {
			field := destElem.Type().Field(i)
			fieldMap[field.Name] = destElem.Field(i).Addr().Interface()
		}

		// 创建一个切片来存储扫描结果
		scanArgs := make([]interface{}, len(columns))
		for i, col := range columns {
			if field, ok := fieldMap[col]; ok {
				scanArgs[i] = field
			} else {
				var dummy interface{}
				scanArgs[i] = &dummy
			}
		}

		// 扫描结果到结构体字段
		err = rows.Scan(scanArgs...)
		if err != nil {
			return err
		}
	} else {
		// 扫描结果到非结构体类型
		err := rows.Scan(dest)
		if err != nil {
			return err
		}
	}

	return nil
}
