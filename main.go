package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"slices"
	"time"

	"github.com/alexflint/go-arg"
	"github.com/bjschafer/zfs-rclone/internal/rclone"
	"github.com/bjschafer/zfs-rclone/internal/schedule"
	"github.com/bjschafer/zfs-rclone/internal/zfs"
)

var args struct {
	Remote                string `arg:"-r,required" help:"Name of already-configured rclone remote"`
	ZpoolName             string `arg:"-p,--zpool-name,required" help:"Name of zpool to backup"`
	ProcessedPropertyName string `arg:"--processed-property-name" default:"backups:wasabi:last_processed" help:"ZFS user property that stores last processed time"`
	SchedulePropertyName  string `arg:"--schedule-property-name" default:"backups:wasabi:schedule" help:"ZFS user property that stores backup schedule"`
	Strategy              string `default:"sync" help:"Rclone strategy to use when backing up [sync, copy, move]"`
	ZfsType               string `arg:"--zfs-type" default:"filesystem" help:"Type of ZFS to backup [filesystem, snap, vol, all]"`
	Verbose               bool   `arg:"-v,--verbose" help:"Increase log level"`
}

func main() {
	p := arg.MustParse(&args)

	validStrategies := []string{"sync", "copy", "move"}
	validZfsTypes := []string{"filesystem", "snap", "vol", "all"}

	if !slices.Contains(validStrategies, args.Strategy) {
		p.Fail(fmt.Sprintf("invalid value for --strategy: %s", args.Strategy))
	}
	if !slices.Contains(validZfsTypes, args.ZfsType) {
		p.Fail(fmt.Sprintf("invalid value for --zfs-type: %s", args.ZfsType))
	}

	if _, err := exec.LookPath("zfs"); errors.Is(err, exec.ErrNotFound) {
		fmt.Fprintln(os.Stderr, "error: couldn't find zfs in $PATH")
		os.Exit(1)
	}
	if _, err := exec.LookPath("rclone"); errors.Is(err, exec.ErrNotFound) {
		fmt.Fprintln(os.Stderr, "error: couldn't find rclone in $PATH")
		os.Exit(1)
	}

	var level slog.Level
	if args.Verbose {
		level = slog.LevelDebug
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	z := zfs.New()
	candidates, err := z.GetScheduledDatasets(ctx, args.SchedulePropertyName, args.ZpoolName, args.ZfsType)
	if err != nil {
		logger.Error("Error getting scheduled datasets", "error", err)
		os.Exit(1)
	}

	for fsName, fsSchedule := range candidates {
		lastProcessed, err := z.GetLastProcessedTime(ctx, args.ProcessedPropertyName, fsName)
		logger.Debug("considering for backup", "fsName", fsName, "fsSchedule", fsSchedule, "lastProcessed", lastProcessed)
		if err != nil {
			logger.Error("Error getting last processed time", "error", err, "fsName", fsName)
			os.Exit(1)
		}
		if !schedule.ShouldProcess(fsSchedule, lastProcessed) {
			logger.Debug("skipping as it's not time yet", "fsName", fsName, "fsSchedule", fsSchedule, "lastProcessed", lastProcessed)
			continue
		}

		r := rclone.New().
			WithCopyLinks().
			WithFastList().
			WithRemote(args.Remote).
			WithStrategy(rclone.Strategy(args.Strategy)).
			WithSyslog()

		logger.Debug("starting rclone", "fsName", fsName)
		err = r.Do(ctx, fsName)
		if err != nil {
			logger.Error("error running rclone", "error", err, "fsName", fsName)
		}

		now := time.Now()
		err = z.SetProcessedTime(ctx, args.ProcessedPropertyName, fsName, &now)
		if err != nil {
			logger.Error("error setting last processed time", "error", err, "fsName", fsName)
		}
	}
}
