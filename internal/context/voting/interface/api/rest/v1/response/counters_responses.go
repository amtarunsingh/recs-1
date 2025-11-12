package response

import "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/counter/entity"

type CountersGroup struct {
	IncomingYes uint32 `json:"incoming_yes" doc:"Incoming yes votes count"`
	IncomingNo  uint32 `json:"incoming_no" doc:"Incoming no votes count"`
	OutgoingYes uint32 `json:"outgoing_yes" doc:"Outgoing yes votes count"`
	OutgoingNo  uint32 `json:"outgoing_no" doc:"Outgoing no votes count"`
}

func NewCountersGroupFromEntity(counters *entity.CountersGroup) CountersGroup {
	return CountersGroup{
		IncomingYes: counters.IncomingYes,
		IncomingNo:  counters.IncomingNo,
		OutgoingYes: counters.OutgoingYes,
		OutgoingNo:  counters.OutgoingNo,
	}
}

type LifetimeCountersGetResponse struct {
	Body CountersGroup
}

func CreateLifetimeCountersGetResponseFromCountersGroup(counters entity.CountersGroup) *LifetimeCountersGetResponse {
	resp := &LifetimeCountersGetResponse{
		Body: NewCountersGroupFromEntity(&counters),
	}
	return resp
}

type HourlyCountersGetResponse struct {
	Body map[uint8]CountersGroup
}

func CreateHourlyCountersGetResponseFromCountersGroup(counters map[uint8]*entity.CountersGroup) *HourlyCountersGetResponse {
	resp := &HourlyCountersGetResponse{
		Body: map[uint8]CountersGroup{},
	}

	for group, counter := range counters {
		resp.Body[group] = NewCountersGroupFromEntity(counter)
	}

	return resp
}
