package rclone

import "context"

type runFunc func(ctx context.Context, cmdName string, args []string) (stdout, stderr []string, err error)

type Rclone struct {
	run      runFunc
	args     []string
	strategy Strategy
	remote   string
}

type Strategy string

var (
	StrategySync Strategy = "sync"
	StrategyMove Strategy = "move"
	StrategyCopy Strategy = "copy"
)

func New() *Rclone {
	// if !os.LookPath("rclone") ...

	return &Rclone{
		run: run,
		args: []string{
			"--one-file-system",
			"--exclude-if-present=.nobackup",
		},
	}
}

func (r *Rclone) WithStrategy(s Strategy) *Rclone {
	r.strategy = s
	return r
}

func (r *Rclone) WithRemote(rem string) *Rclone {
	r.remote = rem
	return r
}

func (r *Rclone) WithFastList() *Rclone {
	r.args = append(r.args, "--fast-list")
	return r
}

func (r *Rclone) WithCopyLinks() *Rclone {
	r.args = append(r.args, "--copy-links")
	return r
}

func (r *Rclone) WithSyslog() *Rclone {
	r.args = append(r.args, "--syslog")
	return r
}

func (r *Rclone) WithBandwidthLimit(limit string) *Rclone {
	// TODO future improvement: parse and validate rclone's weird bandwidth syntax
	r.args = append(r.args, "--bwlimit="+limit)
	return r
}

// Do runs a configured Rclone. The caller is expected to verify other rclone instances aren't runnning
func (r *Rclone) Do(ctx context.Context, fsName string) error {
	args := []string{
		string(r.strategy),
	}
	args = append(args, r.args...)
	args = append(args,
		"/"+fsName,
		r.remote+":"+fsName,
	)
	_, _, err := r.run(ctx, "rclone", args)
	return err
}

func run(ctx context.Context, cmdName string, args []string) (stdout, stderr []string, err error) {

	return nil, nil, nil
}
