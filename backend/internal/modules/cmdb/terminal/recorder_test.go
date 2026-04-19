package terminal

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRecorderWritesCastHeaderAndEvents(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	castPath := filepath.Join(tempDir, "nested", "session.cast")

	recorder, err := NewRecorder(castPath, 132, 43)
	require.NoError(t, err)

	require.NoError(t, recorder.Record("o", "login as: "))
	require.NoError(t, recorder.Record("o", "root\r\n"))
	require.NoError(t, recorder.Close())
	require.NoError(t, recorder.Close())

	payload, err := ParseCastFile(castPath)
	require.NoError(t, err)
	require.NotNil(t, payload)
	require.Equal(t, 132, payload.Width)
	require.Equal(t, 43, payload.Height)
	require.Len(t, payload.Events, 2)
	require.Equal(t, "o", payload.Events[0].Type)
	require.Equal(t, "login as: ", payload.Events[0].Data)
	require.Equal(t, "o", payload.Events[1].Type)
	require.Equal(t, "root\r\n", payload.Events[1].Data)
	require.GreaterOrEqual(t, payload.Events[0].Time, 0.0)
	require.GreaterOrEqual(t, payload.Events[1].Time, payload.Events[0].Time)
}
