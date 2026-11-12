package services

import "errors"

var ErrCascadeRequiredForChildCapabilities = errors.New("cascade deletion required for capability with descendants")

var ErrDeleteImpactChanged = errors.New("the realizations removed by this delete changed since the impact was confirmed")
