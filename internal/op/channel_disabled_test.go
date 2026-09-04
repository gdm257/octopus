package op

import (
	"testing"

	"github.com/bestruirui/octopus/internal/model"
)

func TestChannelGrantAvailability(t *testing.T) {
	cases := []struct {
		name           string
		channelPresent bool
		channelEnabled bool
		keyEnabled     bool
		wantDisabled   bool
		wantGetError   string
	}{
		{"disabled channel", true, false, true, true, "channel 1 is disabled"},
		{"enabled channel", true, true, true, false, ""},
		{"missing channel", false, false, true, false, "channel 1 not found"},
		{"disabled key", true, true, false, false, "channel key 3 is disabled"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resetChannelCaches(t)
			channelGrantCache.Set(4, model.ChannelGrant{ID: 4, ChannelModelID: 2, ChannelKeyID: 3})
			channelModelCache.Set(2, model.ChannelModel{ID: 2, ChannelID: 1})
			channelKeyCache.Set(3, model.ChannelKey{ID: 3, ChannelKeyConfig: model.ChannelKeyConfig{Enabled: tc.keyEnabled}})
			if tc.channelPresent {
				channelCache.Set(1, model.Channel{ID: 1, ChannelConfig: model.ChannelConfig{Enabled: tc.channelEnabled}})
			}

			if got := ChannelGrantDisabled(4); got != tc.wantDisabled {
				t.Errorf("ChannelGrantDisabled = %t, want %t", got, tc.wantDisabled)
			}
			_, err := ChannelGrantGet(4)
			if tc.wantGetError == "" && err != nil {
				t.Errorf("ChannelGrantGet error = %v", err)
			}
			if tc.wantGetError != "" && (err == nil || err.Error() != tc.wantGetError) {
				t.Errorf("ChannelGrantGet error = %v, want %q", err, tc.wantGetError)
			}
		})
	}
}

func resetChannelCaches(t *testing.T) {
	t.Helper()
	channels := channelCache.GetAll()
	keys := channelKeyCache.GetAll()
	models := channelModelCache.GetAll()
	grants := channelGrantCache.GetAll()
	channelCache.Clear()
	channelKeyCache.Clear()
	channelModelCache.Clear()
	channelGrantCache.Clear()
	t.Cleanup(func() {
		channelCache.Clear()
		channelKeyCache.Clear()
		channelModelCache.Clear()
		channelGrantCache.Clear()
		for id, channel := range channels {
			channelCache.Set(id, channel)
		}
		for id, key := range keys {
			channelKeyCache.Set(id, key)
		}
		for id, channelModel := range models {
			channelModelCache.Set(id, channelModel)
		}
		for id, grant := range grants {
			channelGrantCache.Set(id, grant)
		}
	})
}
