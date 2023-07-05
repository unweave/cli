package session

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/unweave/cli/client/clientfakes"
	"github.com/unweave/cli/config"
	"github.com/unweave/unweave/api/types"
)

func TestCreateSession(t *testing.T) {
	config.Config.Project.URI = "test/testo"
	setup := func() (context.Context, *clientfakes.FakeExecer) {
		execer := new(clientfakes.FakeExecer)
		provider := new(clientfakes.FakeProvider)

		uwc.Exec = execer
		uwc.Provider = provider

		ctx := context.Background()
		return ctx, execer
	}

	t.Run(
		"should create session from full params",
		shouldCreateSessionFullParams(setup),
	)

	t.Run(
		"should fail on invalid params",
		shouldFailOnInvalidParams(setup),
	)
}

func shouldCreateSessionFullParams(
	setup func() (context.Context, *clientfakes.FakeExecer),
) func(t *testing.T) {
	return func(t *testing.T) {
		ctx, execer := setup()

		createdExec := &types.Exec{ID: "created-exec-id"}
		execer.CreateReturns(createdExec, nil)

		params := types.ExecCreateParams{
			Name:     "foot-alphabet-house-trailer",
			Provider: "unweave",
			Spec: types.HardwareSpec{
				GPU: types.GPU{
					Type:  "rtx_4000",
					Count: types.HardwareRequestRange{Min: 1, Max: 1},
					RAM:   types.HardwareRequestRange{Min: 4, Max: 4},
				},
				CPU: types.CPU{
					Type:                 "x86_64",
					HardwareRequestRange: types.HardwareRequestRange{Min: 2, Max: 2},
				},
				RAM: types.HardwareRequestRange{Min: 8, Max: 8},
				HDD: types.HardwareRequestRange{Min: 10, Max: 10},
			},
		}

		id, err := Create(ctx, params)
		assert.NoError(t, err)
		assert.Equal(t, id, createdExec.ID)

		_, _, _, gotParams := execer.CreateArgsForCall(0)
		assert.Equal(t, params, gotParams)
	}
}

func shouldFailOnInvalidParams(
	setup func() (context.Context, *clientfakes.FakeExecer),
) func(t *testing.T) {
	return func(t *testing.T) {
		ctx, execer := setup()

		createdExec := &types.Exec{ID: "created-exec-id"}
		execer.CreateReturns(createdExec, nil)

		params := types.ExecCreateParams{
			Name:     "foot-alphabet-house-trailer",
			Provider: "unweave",
			Spec: types.HardwareSpec{
				GPU: types.GPU{},
				CPU: types.CPU{},
			},
		}

		_, err := Create(ctx, params)

		var wantErr *types.Error
		assert.Error(t, err)
		assert.ErrorAs(t, err, &wantErr)

		assert.Equal(t, execer.CreateCallCount(), 0)
	}
}
