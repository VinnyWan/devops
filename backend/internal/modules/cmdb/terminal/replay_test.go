package terminal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseCastFile(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	castPath := filepath.Join(tempDir, "session.cast")
	castContent := `{"version": 2, "width": 120, "height": 40, "timestamp": 1713254400}
[0.01, "o", "login as: "]
[0.50, "o", "root\r\n"]
[1.25, "o", "Last login\r\n"]
`
	require.NoError(t, os.WriteFile(castPath, []byte(castContent), 0o600))

	payload, err := ParseCastFile(castPath)
	require.NoError(t, err)
	require.NotNil(t, payload)
	require.Equal(t, 120, payload.Width)
	require.Equal(t, 40, payload.Height)
	require.Len(t, payload.Events, 3)
	require.Equal(t, 0.01, payload.Events[0].Time)
	require.Equal(t, "o", payload.Events[0].Type)
	require.Equal(t, "login as: ", payload.Events[0].Data)
	require.Equal(t, 1.25, payload.Events[2].Time)
	require.Equal(t, "Last login\r\n", payload.Events[2].Data)
}
