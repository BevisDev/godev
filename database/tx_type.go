package database

import (
	"database/sql"
	"strconv"
)

type IsoLevel int

const (
	Default IsoLevel = iota
	ReadUncommitted
	ReadCommitted
	WriteCommitted
	RepeatableRead
	Snapshot
	Serializable
	Linearizable
)

// String returns the name of the transaction isolation level.
func (i IsoLevel) String() string {
	switch i {
	case Default:
		return "Default"
	case ReadUncommitted:
		return "Read Uncommitted"
	case ReadCommitted:
		return "Read Committed"
	case WriteCommitted:
		return "Write Committed"
	case RepeatableRead:
		return "Repeatable Read"
	case Snapshot:
		return "Snapshot"
	case Serializable:
		return "Serializable"
	case Linearizable:
		return "Linearizable"
	default:
		return "IsolationLevel(" + strconv.Itoa(int(i)) + ")"
	}
}

var toSqlIsoLevel = map[IsoLevel]sql.IsolationLevel{
	Default:         sql.LevelDefault,
	ReadUncommitted: sql.LevelReadUncommitted,
	ReadCommitted:   sql.LevelReadCommitted,
	WriteCommitted:  sql.LevelWriteCommitted,
	RepeatableRead:  sql.LevelRepeatableRead,
	Snapshot:        sql.LevelSnapshot,
	Serializable:    sql.LevelSerializable,
	Linearizable:    sql.LevelLinearizable,
}

func (i IsoLevel) ToSqlIsoLevel() sql.IsolationLevel {
	return toSqlIsoLevel[i]
}
