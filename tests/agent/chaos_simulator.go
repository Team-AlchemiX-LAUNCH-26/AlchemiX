package agenttest

import (
	"fmt"
	"sort"
)

type LinkState struct {
	LinkID                string
	PlanetA               string
	PlanetB               string
	CapacityUnits         float64
	CurrentLoad           float64
	LoadRatio             float64
	SelfReportedLatencyMS *float64
	TrafficShare          float64
	Status                string
	Jammed                bool
	DropEveryN            int
}

type Snapshot struct {
	Tick  int64
	Links map[string]LinkState
}

type Mutation func(*Snapshot)

type Scenario struct {
	Name      string
	Snapshots []Snapshot
}

type ScenarioBuilder struct {
	name      string
	base      Snapshot
	mutations map[int64][]Mutation
	lastTick  int64
}

func NewScenario(name string, base Snapshot) *ScenarioBuilder {
	return &ScenarioBuilder{
		name:      name,
		base:      cloneSnapshot(base),
		mutations: make(map[int64][]Mutation),
		lastTick:  base.Tick,
	}
}

func (b *ScenarioBuilder) AtTick(tick int64, mutations ...Mutation) *ScenarioBuilder {
	b.mutations[tick] = append(b.mutations[tick], mutations...)
	if tick > b.lastTick {
		b.lastTick = tick
	}
	return b
}

func (b *ScenarioBuilder) UntilTick(tick int64) *ScenarioBuilder {
	if tick > b.lastTick {
		b.lastTick = tick
	}
	return b
}

func (b *ScenarioBuilder) Build() Scenario {
	current := cloneSnapshot(b.base)
	result := Scenario{Name: b.name}

	for tick := b.base.Tick; tick <= b.lastTick; tick++ {
		current.Tick = tick
		for _, mutation := range b.mutations[tick] {
			mutation(&current)
		}
		result.Snapshots = append(result.Snapshots, cloneSnapshot(current))
	}

	return result
}

func Saturate(linkID string) Mutation {
	return func(snapshot *Snapshot) {
		link, ok := snapshot.Links[linkID]
		if !ok {
			panic(fmt.Sprintf("unknown link %q", linkID))
		}
		link.LoadRatio = 0.95
		link.CurrentLoad = link.CapacityUnits * link.LoadRatio
		link.Status = "saturated"
		link.SelfReportedLatencyMS = nil
		snapshot.Links[linkID] = link
	}
}

func Restore(linkID string, loadRatio float64, latencyMS float64) Mutation {
	return func(snapshot *Snapshot) {
		link, ok := snapshot.Links[linkID]
		if !ok {
			panic(fmt.Sprintf("unknown link %q", linkID))
		}
		link.LoadRatio = loadRatio
		link.CurrentLoad = link.CapacityUnits * loadRatio
		link.Status = "ok"
		link.Jammed = false
		link.SelfReportedLatencyMS = floatPointer(latencyMS)
		snapshot.Links[linkID] = link
	}
}

func Jam(linkID string) Mutation {
	return func(snapshot *Snapshot) {
		link, ok := snapshot.Links[linkID]
		if !ok {
			panic(fmt.Sprintf("unknown link %q", linkID))
		}
		link.Jammed = true
		snapshot.Links[linkID] = link
	}
}

func SpoofLatency(linkID string, latencyMS float64) Mutation {
	return func(snapshot *Snapshot) {
		link, ok := snapshot.Links[linkID]
		if !ok {
			panic(fmt.Sprintf("unknown link %q", linkID))
		}
		link.SelfReportedLatencyMS = floatPointer(latencyMS)
		snapshot.Links[linkID] = link
	}
}

func SetTrafficShare(linkID string, share float64) Mutation {
	return func(snapshot *Snapshot) {
		link, ok := snapshot.Links[linkID]
		if !ok {
			panic(fmt.Sprintf("unknown link %q", linkID))
		}
		link.TrafficShare = share
		snapshot.Links[linkID] = link
	}
}

func Grayhole(linkID string, dropEveryN int) Mutation {
	return func(snapshot *Snapshot) {
		link, ok := snapshot.Links[linkID]
		if !ok {
			panic(fmt.Sprintf("unknown link %q", linkID))
		}
		link.DropEveryN = dropEveryN
		snapshot.Links[linkID] = link
	}
}

func UnknownStatus(linkID, status string) Mutation {
	return func(snapshot *Snapshot) {
		link, ok := snapshot.Links[linkID]
		if !ok {
			panic(fmt.Sprintf("unknown link %q", linkID))
		}
		link.Status = status
		snapshot.Links[linkID] = link
	}
}

func RemoveLink(linkID string) Mutation {
	return func(snapshot *Snapshot) {
		delete(snapshot.Links, linkID)
	}
}

func RepeatTick() Mutation {
	return func(snapshot *Snapshot) {
		snapshot.Tick--
	}
}

func SortedLinkIDs(snapshot Snapshot) []string {
	ids := make([]string, 0, len(snapshot.Links))
	for id := range snapshot.Links {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func cloneSnapshot(source Snapshot) Snapshot {
	target := Snapshot{
		Tick:  source.Tick,
		Links: make(map[string]LinkState, len(source.Links)),
	}
	for id, link := range source.Links {
		if link.SelfReportedLatencyMS != nil {
			value := *link.SelfReportedLatencyMS
			link.SelfReportedLatencyMS = &value
		}
		target.Links[id] = link
	}
	return target
}

func floatPointer(value float64) *float64 {
	return &value
}
