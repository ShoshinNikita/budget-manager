// Package db contains common entities (errors, models and etc). All DB implementations have to use them
package db

type Type int

const (
	Unknown = iota
	Postgres
	Sqlite3
)

func (t Type) String() string {
	switch t {
	case Postgres:
		return "postgres"
	case Sqlite3:
		return "sqlite"
	default:
		return "unknown"
	}
}

func (t *Type) UnmarshalText(text []byte) error {
	switch string(text) {
	case "postgres", "postgresql":
		*t = Postgres
	case "sqlite", "sqlite3":
		*t = Sqlite3
	default:
		*t = Unknown
	}
	return nil
}
