package logger

func Error(err error) Field {
	return Field{
		Key: "error",
		Val: err,
	}
}

func String(key string, val string) Field {
	return Field{
		Key: key,
		Val: val,
	}
}

func Bool(key string, val bool) Field {
	return Field{
		Key: key,
		Val: val,
	}
}

func Int(key string, val int) Field {
	return Field{
		Key: key,
		Val: val,
	}
}

func Int64(key string, val int64) Field {
	return Field{
		Key: key,
		Val: val,
	}
}
