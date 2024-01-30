package ratelimit

import _ "embed"

//go:embed slide_window.lua
var slideWindowLuaScript string

//go:embed token_bucket.lua
var tokenBucketLuaScript string
