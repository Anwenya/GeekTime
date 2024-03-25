package migrator

type Entity interface {
	// ID 返回id
	ID() int64
	// CompareTo 比较
	CompareTo(dst Entity) bool
}
