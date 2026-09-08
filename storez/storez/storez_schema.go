package storez

type StorezSchemaColumn struct {
	Name    string
	Indexed bool
}

type StorezSchemaTable struct {
	Name    string
	Columns map[string]*StorezSchemaColumn
}

type StorezSchema struct {
	Tables map[string]*StorezSchemaTable
}

func CreateSchema() *StorezSchema {
	return &StorezSchema{Tables: map[string]*StorezSchemaTable{}}
}

func (me *StorezSchema) Index(tableName string, columnNames ...string) {
	table, found := me.Tables[tableName]
	if !found {
		table = &StorezSchemaTable{Name: tableName, Columns: map[string]*StorezSchemaColumn{}}
		me.Tables[tableName] = table
	}
	for _, columnName := range columnNames {
		column, found := table.Columns[columnName]
		if !found {
			column = &StorezSchemaColumn{Name: columnName}
			table.Columns[columnName] = column
		}
		column.Indexed = true
	}
}

func (me *StorezSchema) GetTable(tableName string) *StorezSchemaTable {
	table, found := me.Tables[tableName]
	if !found {
		return &StorezSchemaTable{Name: tableName}
	}
	return table
}

func (me *StorezSchemaTable) GetColumn(columnName string) *StorezSchemaColumn {
	table, found := me.Columns[columnName]
	if !found {
		return &StorezSchemaColumn{Name: columnName}
	}
	return table
}
