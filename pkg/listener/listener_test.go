package listener

import (
	"testing"
	"time"

	"github.com/els0r/dynip-ng/pkg/cfg"
	"github.com/els0r/dynip-ng/pkg/listener/state"
	"github.com/els0r/dynip-ng/pkg/update"
)

type mockUpdater struct{}

func (m *mockUpdater) Update(ip string) error { return nil }
func (m *mockUpdater) Name() string           { return "mock updater" }

func TestListener(t *testing.T) {
	var tests = []struct {
		name   string
		config *cfg.ListenConfig
	}{
		{
			"test run and stop (long)",
			&cfg.ListenConfig{
				Interval: 1,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			config := test.config

			state := state.NewInMemory()
			mu := &mockUpdater{}

			// create listener
			l, err := New(config, state, []update.Updater{
				// line up all different updaters
				mu,
			}...)
			if err != nil {
				t.Fatalf("failed to create listener: %s", err)
			}

			// and run it
			stop := l.Run()

			// wait for at least one ticker cycle to complete
			// TODO: this slows down the test significatly, since the minimum
			// duration is at 1 minute at the moment. Change config to accept seconds
			time.Sleep(time.Duration(config.Interval)*time.Minute + 5*time.Second)

			// check if the listener can be stopped
			stop <- struct{}{}
		})
	}
}
