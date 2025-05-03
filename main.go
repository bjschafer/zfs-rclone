package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"slices"

	"github.com/alexflint/go-arg"
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

	var level slog.Level
	if args.Verbose {
		level = slog.LevelDebug
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	logger := slog.New(handler)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
}
