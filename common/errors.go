package common

import (
	"github.com/pkg/errors"
)

var NilInputErr error = errors.New("nil input")
var ParamInvalidErr error = errors.New("invalid parameter")
