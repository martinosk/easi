package services

import "errors"

var ErrCascadeRequiredForChildCapabilities = errors.New("cascade deletion required for capability with descendants")
