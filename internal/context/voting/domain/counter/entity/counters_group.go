package entity

import (
	sharedValueObject "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/sharedkernel/valueobject"
)

type CountersGroup struct {
	ActiveUserKey     sharedValueObject.ActiveUserKey
	HourUnixTimestamp int32
	IncomingYes       uint32
	IncomingNo        uint32
	OutgoingYes       uint32
	OutgoingNo        uint32
}
