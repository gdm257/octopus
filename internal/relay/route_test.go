package relay

import (
	"testing"

	"github.com/bestruirui/octopus/internal/model"
)

func TestPickGroupItem(t *testing.T) {
	cases := []struct {
		name  string
		group model.Group
		skip  func(model.GroupItem) bool
		want  int
	}{
		{
			name:  "failover skips disabled member",
			group: model.Group{ID: 1, Mode: model.GroupModeFailover, Items: []model.GroupItem{{ID: 11}, {ID: 12}}},
			skip:  func(item model.GroupItem) bool { return item.ID == 11 },
			want:  12,
		},
		{
			name:  "manual does not fall back",
			group: model.Group{ID: 2, Mode: model.GroupModeManual, ActiveItemID: 21, Items: []model.GroupItem{{ID: 21}, {ID: 22}}},
			skip:  func(item model.GroupItem) bool { return item.ID == 21 },
		},
		{
			name:  "all members skipped",
			group: model.Group{ID: 3, Mode: model.GroupModeFailover, Items: []model.GroupItem{{ID: 31}, {ID: 32}}},
			skip:  func(model.GroupItem) bool { return true },
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ResetRouteState(tc.group.ID)
			if got := pickGroupItem(tc.group, tc.skip); got.ID != tc.want {
				t.Errorf("pickGroupItem ID = %d, want %d", got.ID, tc.want)
			}
		})
	}
}

func TestPickGroupItemClearsSkippedRouteState(t *testing.T) {
	group := model.Group{ID: 4, Mode: model.GroupModeFailover, Items: []model.GroupItem{{ID: 41}, {ID: 42}}}
	ResetRouteState(group.ID)
	routeMu.Lock()
	routes[group.ID] = &RouteState{GroupID: group.ID, CurrentItemID: 41, ProbeItemID: 41, Cooldowns: map[int]int64{}}
	routeMu.Unlock()

	if got := pickGroupItem(group, func(item model.GroupItem) bool { return item.ID == 41 }); got.ID != 42 {
		t.Errorf("pickGroupItem ID = %d, want 42", got.ID)
	}
	state := RouteStateOf(group)
	if state.CurrentItemID != 42 || state.ProbeItemID != 0 {
		t.Errorf("RouteStateOf = %#v, want current 42 and no probe", state)
	}
}

func TestAllItemsSkipped(t *testing.T) {
	cases := []struct {
		name  string
		group model.Group
		skip  func(model.GroupItem) bool
		want  bool
	}{
		{
			name:  "failover all skipped",
			group: model.Group{Mode: model.GroupModeFailover, Items: []model.GroupItem{{ID: 1}, {ID: 2}}},
			skip:  func(model.GroupItem) bool { return true },
			want:  true,
		},
		{
			name:  "failover member remains",
			group: model.Group{Mode: model.GroupModeFailover, Items: []model.GroupItem{{ID: 1}, {ID: 2}}},
			skip:  func(item model.GroupItem) bool { return item.ID == 1 },
		},
		{
			name:  "empty failover group",
			group: model.Group{Mode: model.GroupModeFailover},
			skip:  func(model.GroupItem) bool { return true },
		},
		{
			name:  "manual active member skipped",
			group: model.Group{Mode: model.GroupModeManual, ActiveItemID: 2, Items: []model.GroupItem{{ID: 1}, {ID: 2}}},
			skip:  func(item model.GroupItem) bool { return item.ID == 2 },
			want:  true,
		},
		{
			name:  "manual active member absent",
			group: model.Group{Mode: model.GroupModeManual, ActiveItemID: 3, Items: []model.GroupItem{{ID: 1}, {ID: 2}}},
			skip:  func(model.GroupItem) bool { return true },
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := allItemsSkipped(tc.group, tc.skip); got != tc.want {
				t.Errorf("allItemsSkipped = %t, want %t", got, tc.want)
			}
		})
	}
}
