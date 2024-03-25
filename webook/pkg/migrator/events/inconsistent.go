package events

const (
	// InconsistentEventTypeTargetMissing 目标数据缺失
	InconsistentEventTypeTargetMissing = "target_missing"
	// InconsistentEventTypeBaseMissing 源数据缺失
	InconsistentEventTypeBaseMissing = "base_missing"
	// InconsistentEventTypeNEQ 不相等
	InconsistentEventTypeNEQ = "neq"
)

type InconsistentEvent struct {
	ID int64
	// SRC 以源表为准
	// DST 以目标表为准
	Direction string
	// 不一致的原因
	Type string
}
