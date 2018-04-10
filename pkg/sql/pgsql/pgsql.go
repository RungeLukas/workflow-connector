package pgsql

import (
	"context"
	"database/sql"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/signavio/workflow-connector/pkg/config"
	sqlBackend "github.com/signavio/workflow-connector/pkg/sql"
)

type lastId struct {
	id int64
}

func NewPgsqlBackend(cfg *config.Config, router *mux.Router) (b *sqlBackend.Backend) {
	b = sqlBackend.NewBackend(cfg, router)
	b.ConvertDBSpecificDataType = convertFromPgsqlDataType
	b.Queries = map[string]string{
		"GetSingleAsOption":                "SELECT id, %s FROM %s WHERE id = $1",
		"GetCollection":                    "SELECT * FROM %s",
		"GetCollectionAsOptions":           "SELECT id, %s FROM %s",
		"GetCollectionAsOptionsFilterable": "SELECT id, %s FROM %s WHERE CAST (%s AS TEXT) LIKE $1",
		"GetTableSchema":                   "SELECT * FROM %s LIMIT 1",
	}
	b.Templates = map[string]string{
		"GetTableWithRelationshipsSchema": "SELECT * FROM {{.TableName}} AS _{{.TableName}}" +
			"{{range .Relations}}" +
			" LEFT JOIN {{.Relationship.WithTable}}" +
			" ON {{.Relationship.WithTable}}.{{.Relationship.ForeignKey}}" +
			" = _{{$.TableName}}.id{{end}} LIMIT 1",
		"GetSingleWithRelationships": "SELECT * FROM {{.TableName}} AS _{{.TableName}}" +
			"{{range .Relations}}" +
			" LEFT JOIN {{.Relationship.WithTable}}" +
			" ON {{.Relationship.WithTable}}.{{.Relationship.ForeignKey}}" +
			" = _{{$.TableName}}.id{{end}}" +
			" WHERE _{{$.TableName}}.id = $1",
		"UpdateSingle": "UPDATE {{.Table}} SET {{.ColumnNames | head}}" +
			" = $1{{range $index, $element := .ColumnNames | tail}}," +
			" {{$element}} = ${{(add2 $index)}}{{end}}" +
			" WHERE id = ${{(lenPlus1 .ColumnNames)}}",
		"CreateSingle": "INSERT INTO {{.Table}}({{.ColumnNames | head}}" +
			"{{range .ColumnNames | tail}}, {{.}}{{end}}) VALUES($1{{range $index," +
			" $element := .ColumnNames | tail}}, ${{$index | add2}}{{end}}) RETURNING id",
	}
	b.TransactDirectly = execContextDirectly
	b.TransactWithinTx = execContextWithinTx
	return b
}

func (l *lastId) LastInsertId() (int64, error) {
	return l.id, nil
}

func (l *lastId) RowsAffected() (int64, error) {
	return 0, nil
}

func execContextDirectly(ctx context.Context, db *sql.DB, query string, args ...interface{}) (result sql.Result, err error) {
	var id int64
	if err = db.QueryRowContext(ctx, query, args...).Scan(&id); err != nil {
		return nil, err
	}
	result = &lastId{id}
	return result, nil
}
func execContextWithinTx(ctx context.Context, tx *sql.Tx, query string, args ...interface{}) (result sql.Result, err error) {
	var id int64
	if err = tx.QueryRowContext(ctx, query, args...).Scan(&id); err != nil {
		return nil, err
	}
	result = &lastId{id}
	return result, nil
}
func convertFromPgsqlDataType(fieldDataType string) interface{} {
	switch fieldDataType {
	// Text data types
	case "CHAR":
		return &sql.NullString{}
	case "VARCHAR":
		return &sql.NullString{}
	case "TEXT":
		return &sql.NullString{}
	case "BYTEA":
		return &sql.NullString{}
	// Number data types
	case "INT2":
		return &sql.NullInt64{}
	case "INT4":
		return &sql.NullInt64{}
	case "INT8":
		return &sql.NullInt64{}
	case "NUMERIC":
		return &sql.NullFloat64{}
	case "MONEY":
		return &sql.NullFloat64{}
	// Date data types
	case "TIMESTAMP":
		return &sqlBackend.NullTime{}
	case "TIMESTAMPTZ":
		return &sqlBackend.NullTime{}
	case "DATE":
		return &sqlBackend.NullTime{}
	case "TIME":
		return &sqlBackend.NullTime{}
	case "TIMETZ":
		return &sqlBackend.NullTime{}
	// Other data types
	case "BOOL":
		return &sql.NullBool{}
	default:
		return &sql.NullString{}
	}
}
