package scanner

type SchemaInfo struct {
	Name   string
	Tables []TableInfo
}

type TableInfo struct {
	Schema      string
	Name        string
	ColumnCount int
	RowCount    int64
}

type ColumnInfo struct {
	Name       string
	DataType   string
	IsNullable bool
}

type DatabaseInfo struct {
	Name         string
	Schemas      []SchemaInfo
	TotalSchemas int
	TotalTables  int
}
