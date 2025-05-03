package zfs

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetScheduledDatasets(t *testing.T) {
	tests := []struct {
		name      string
		runStdout []string
		runStderr []string
		want      map[string]string
		wantErr   error
	}{
		{
			name: "no scheduled datasets",
			runStdout: []string{
				"tank\t-",
				"tank/Example\t-",
				"tank/Example/Archive\t-",
			},
			want: map[string]string{},
		},
		{
			name: "all scheduled datasets",
			runStdout: []string{
				"tank\tdaily",
				"tank/Example\thourly",
				"tank/Example/Archive\tmonthly",
			},
			want: map[string]string{
				"tank":                 "daily",
				"tank/Example":         "hourly",
				"tank/Example/Archive": "monthly",
			},
		},
		{
			name: "some scheduled datasets",
			runStdout: []string{
				"tank\tdaily",
				"tank/Example\t-",
				"tank/Example/Archive\tmonthly",
			},
			want: map[string]string{
				"tank":                 "daily",
				"tank/Example/Archive": "monthly",
			},
		},
		{
			name: "unrecognized schedule",
			runStdout: []string{
				"tank\tquasimonthly",
				"tank/Example\tminutely",
				"tank/Example/Archive\tmonthly",
			},
			want: map[string]string{
				"tank/Example/Archive": "monthly",
			},
		},
		{
			name: "unexpected output",
			runStdout: []string{
				"hello\tthere",
				"Lol you put something besides ZFS in your path",
			},
			wantErr: errors.New("unexpected output on line 2: Lol you put something besides ZFS in your path"),
		},
	}

	for _, tt := range tests {
		handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
		logger := slog.New(handler)
		slog.SetDefault(logger)

		t.Run(tt.name, func(t *testing.T) {
			z := &ZFS{
				run: func(ctx context.Context, cmdName string, args []string) (stdout []string, stderr []string, err error) {
					return tt.runStdout, tt.runStderr, nil
				},
			}

			got, gotErr := z.GetScheduledDatasets(context.Background(), "example", "tank", "all")
			if tt.wantErr != nil && assert.Error(t, gotErr) {
				assert.Equal(t, tt.wantErr, gotErr)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func timeFor(t *testing.T, rfc3339 string) *time.Time {
	t.Helper()

	tm, err := time.Parse(time.RFC3339, rfc3339)
	if err != nil {
		t.Error(err)
	}
	tm = tm.UTC()
	return &tm
}

func TestGetLastProcesedTime(t *testing.T) {
	tests := []struct {
		name      string
		runStdout []string
		runStderr []string
		want      *time.Time
		wantErr   error
	}{
		{
			name:    "no output",
			wantErr: errors.New("unexpected output: []"),
		},
		{
			name:      "never processed",
			runStdout: []string{"-"},
			want:      &time.Time{},
		},
		{
			name:      "valid timestamp",
			runStdout: []string{"1746107062"},
			want:      timeFor(t, "2025-05-01T08:44:22-05:00"),
		},
		{
			name:      "unexpected output",
			runStdout: []string{"zfs properties", "foo\t-"},
			wantErr:   errors.New("unexpected output: [zfs properties foo\t-]"),
		},
	}

	for _, tt := range tests {
		handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
		logger := slog.New(handler)
		slog.SetDefault(logger)

		t.Run(tt.name, func(t *testing.T) {
			z := &ZFS{
				run: func(ctx context.Context, cmdName string, args []string) (stdout []string, stderr []string, err error) {
					return tt.runStdout, tt.runStderr, nil
				},
			}

			got, gotErr := z.GetLastProcessedTime(context.Background(), "example", "tank")
			if tt.wantErr != nil && assert.Error(t, gotErr) {
				assert.Equal(t, tt.wantErr, gotErr)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
