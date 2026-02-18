package udisks

import "errors"

var ErrInvalidDrive = errors.New("invalid drive")
var ErrDriveNotFound = errors.New("drive not found")
var ErrUnmountFailed = errors.New("unmount failed")
var ErrLockingFailed = errors.New("locking failed")
var ErrPowerOffNotSupported = errors.New("power off not supported for this drive")
var ErrInvalidPropertyFormat = errors.New("invalid property format")
