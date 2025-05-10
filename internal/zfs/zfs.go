package zfs

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"strconv"
	"strings"
	"time"

	schedules "github.com/bjschafer/zfs-rclone/internal/schedule"
)

type runFunc func(ctx context.Context, cmdName string, args []string) (stdout, stderr []string, err error)

type ZFS struct {
	run runFunc
}

func New() *ZFS {
	// if !os.LookPath("zfs") ...
	return &ZFS{
		run: run,
	}
}

func (z *ZFS) GetScheduledDatasets(ctx context.Context, schedulePropertyName, zpool, zfsType string) (map[string]string, error) {
	scheduledDatasets := make(map[string]string)

	// zfs does not seem to support long options
	args := []string{
		"get",
		"-H",                 // make more parseable; omit headers and separate fields by exactly one tab.
		"-oname,value",       // get only the name and value columns
		"-r",                 // recursively display for any children
		schedulePropertyName, // property to get
		zpool,                // name of zpool (or subpath therein)
		"-t" + zfsType,       // type to display
	}
	stdout, stderr, err := z.run(ctx, "zfs", args)
	if err != nil {
		slog.Error("failed listing scheduled datasets", "stdout", stdout, "stderr", stderr)
		return nil, fmt.Errorf("failed listing scheduled datasets: %w", err)
	}

	for i, line := range stdout {
		parts := strings.Split(line, "\t")
		if len(parts) != 2 {
			return nil, fmt.Errorf("unexpected output on line %d: %s", i+1, line)
		}

		name := parts[0]
		value := parts[1]

		if value == "-" {
			slog.Debug("skipping unscheduled dataset", "dataset", name)
			continue
		}

		if !schedules.IsValid(value) {
			slog.Warn("invalid schedule set; skipping", "schedule", value, "fsname", name)
			continue
		}
		scheduledDatasets[name] = value
	}

	return scheduledDatasets, nil
}

func (z *ZFS) GetLastProcessedTime(ctx context.Context, processedPropertyName, fsName string) (*time.Time, error) {
	args := []string{
		"get",
		"-H",                  // make more parseable; omit headers and separate fields by exactly one tab.
		"-ovalue",             // get only the value column
		processedPropertyName, // property to get
		fsName,                // name of filesystem or zvol or...
	}
	stdout, stderr, err := z.run(ctx, "zfs", args)
	if err != nil {
		slog.Error("failed getting last processed time", "dataset", fsName, "stdout", stdout, "stderr", stderr)
		return nil, fmt.Errorf("failed getting last processed time: %w", err)
	}

	// stdout should be exactly one line containing either a unix timestamp, or the "-" character.
	if len(stdout) != 1 {
		slog.Error("unexpected output for last processed time", "dataset", fsName, "stdout", stdout, "stderr", stderr)
		return nil, fmt.Errorf("unexpected output: %v", stdout)
	}

	result := strings.TrimSpace(stdout[0])

	// if not set, assume it's never been processed
	if result == "-" {
		return &time.Time{}, nil
	}

	i, err := strconv.ParseInt(result, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse %s as time: %w", result, err)
	}
	t := time.Unix(i, 0).UTC()
	return &t, nil
}

func (z *ZFS) SetProcessedTime(ctx context.Context, processedPropertyName, fsName string, val *time.Time) error {
	args := []string{
		"set",
		fmt.Sprintf("%s=%s", processedPropertyName, val),
		fsName,
	}
	_, _, err := z.run(ctx, "zfs", args)
	return err
}

// run runs a command and converts its stdout and stderr streams to newline-delimited string slices
func run(ctx context.Context, cmdName string, args []string) (stdout, stderr []string, err error) {
	stdoutWriter := new(bytes.Buffer)
	stderrWriter := new(bytes.Buffer)

	slog.Debug("Running command", "command", cmdName, "args", args)
	cmd := exec.CommandContext(ctx, cmdName, args...)
	cmd.Stdout = stdoutWriter
	cmd.Stderr = stderrWriter

	err = cmd.Run()
	if err != nil {
		return nil, nil, err
	}

	stdout = strings.Split(strings.TrimSpace(stdoutWriter.String()), "\n")
	stderr = strings.Split(strings.TrimSpace(stderrWriter.String()), "\n")

	return stdout, stderr, nil
}
