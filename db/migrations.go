package db

import (
	"github.com/forscht/ddrv/pkg/migrate"
)

var migrations = []migrate.Migration{
	{
		ID: 1,
		Up: migrate.Queries([]string{
			fsTable,
			nodeTable,
			fsParentIdx,
			fsNameIdx,
			rootInsert,
			statFunction,
			lsFunction,
			treeFunction,
			touchFunction,
			mkdirFunction,
			mvFunction,
			rmFunction,
			resetFunction,
			parserootFunction,
			validnameFunction,
			sanitizeFPath,
			parseSizeFunction,
			basenameFunction,
			dirnameFunction,
		}),
		Down: migrate.Queries([]string{dropFs}),
	},
}
