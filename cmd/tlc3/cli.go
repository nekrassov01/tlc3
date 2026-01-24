package main

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"slices"
	"strconv"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/nekrassov01/logger/log"
	"github.com/nekrassov01/tlc3"
	"github.com/urfave/cli/v3"
)

const (
	name  = "tlc3"
	label = "TLC3"
)

var (
	logger = &log.Logger{}
)

func newCmd(w, ew io.Writer) *cli.Command {
	logger = log.NewLogger(log.NewCLIHandler(io.Discard))

	loglevel := &cli.StringFlag{
		Name:    "log-level",
		Aliases: []string{"l"},
		Usage:   "set log level",
		Sources: cli.EnvVars(label + "_LOG_LEVEL"),
		Value:   slog.LevelInfo.String(),
	}

	addr := &cli.StringSliceFlag{
		Name:    "address",
		Aliases: []string{"a"},
		Usage:   "domain:port separated by commas",
	}

	file := &cli.StringFlag{
		Name:    "file",
		Aliases: []string{"f"},
		Usage:   "path to newline-delimited list of addresses",
	}

	output := &cli.StringFlag{
		Name:    "output",
		Aliases: []string{"o"},
		Usage:   "set output type",
		Sources: cli.EnvVars(label + "_OUTPUT_TYPE"),
		Value:   tlc3.OutputTypeCompressedText.String(),
	}

	timeout := &cli.DurationFlag{
		Name:    "timeout",
		Aliases: []string{"t"},
		Usage:   "set network timeout duration",
		Value:   5 * time.Second,
		Sources: cli.EnvVars(label + "_TIMEOUT"),
	}

	insecure := &cli.BoolFlag{
		Name:    "insecure",
		Aliases: []string{"i"},
		Usage:   "skip verification of the cert chain and host name",
		Value:   false,
	}

	static := &cli.BoolFlag{
		Name:    "static",
		Aliases: []string{"s"},
		Usage:   "hide fields related to the current time in table output",
		Value:   false,
	}

	timeZone := &cli.StringFlag{
		Name:    "timezone",
		Aliases: []string{"z"},
		Usage:   "time zone for datetime fields",
		Value:   "Local",
		Sources: cli.EnvVars(label + "_TIMEZONE"),
	}

	before := func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
		// parse log level
		var level slog.Level
		if err := level.UnmarshalText([]byte(cmd.String(loglevel.Name))); err != nil {
			level = slog.LevelInfo
		}

		// set logger options
		s := log.Style2()
		s.Caller.Fullpath = true
		withLevel := log.WithLevel(level)
		withCaller := log.WithCaller(level <= slog.LevelDebug)
		withStyle := log.WithStyle(s)

		// create logger for application
		logger = log.NewLogger(log.NewCLIHandler(ew,
			log.WithLabel("TLC3:"),
			log.WithTime(true),
			withLevel,
			withCaller,
			withStyle,
		))

		if cmd.IsSet(addr.Name) && cmd.IsSet(file.Name) {
			return nil, fmt.Errorf("cannot be used together %s and %s", addr.Name, file.Name)
		}

		if cmd.Bool(insecure.Name) {
			if err := confirmInsecure(); err != nil {
				return nil, err
			}
		}

		return ctx, nil
	}

	action := func(ctx context.Context, cmd *cli.Command) error {
		// get addresses
		var addresses []string
		var err error
		if cmd.IsSet(addr.Name) {
			addresses = cmd.StringSlice(addr.Name)
		}
		if cmd.IsSet(file.Name) {
			addresses, err = tlc3.GetAddressesFromFile(cmd.String(file.Name))
			if err != nil {
				return err
			}
		}
		if len(addresses) == 0 {
			return errors.New("cannot receive addresses")
		}

		// load location
		tz := cmd.String(timeZone.Name)
		loc, err := time.LoadLocation(tz)
		if err != nil {
			return fmt.Errorf("cannot load timezone %q", tz)
		}

		// get certificate informations
		logger.Info("getting certificate informations...")
		infos, err := tlc3.GetCerts(ctx, addresses, cmd.Duration(timeout.Name), cmd.Bool(insecure.Name), loc)
		if err != nil {
			return err
		}
		slices.SortFunc(infos, func(a, b *tlc3.CertInfo) int {
			return cmp.Compare(a.DomainName, b.DomainName)
		})

		// parse output type passed as string
		outputType, err := tlc3.ParseOutputType(cmd.String(output.Name))
		if err != nil {
			return err
		}

		// create renderer and render output
		ren := tlc3.NewRenderer(w, infos, outputType, cmd.Bool(static.Name))
		if err := ren.Render(); err != nil {
			return err
		}

		// done message
		logger.Info("completed")
		return nil
	}

	return &cli.Command{
		Name:                  name,
		Version:               getVersion(),
		Usage:                 "TLS cert checker CLI",
		Description:           "CLI application for checking TLS certificate information",
		HideHelpCommand:       true,
		EnableShellCompletion: true,
		Writer:                w,
		ErrWriter:             ew,
		Before:                before,
		Action:                action,
		Flags:                 []cli.Flag{loglevel, addr, file, output, timeout, insecure, static, timeZone},
	}
}

// confirmInsecure prompts the user to confirm the use of the insecure flag.
func confirmInsecure() error {
	ci, _ := strconv.ParseBool(os.Getenv(label + "_NON_INTERACTIVE"))
	if ci {
		return nil
	}
	prompt := promptui.Prompt{
		Label:     "[WARNING] insecure flag skips verification of the certificate chain and hostname. skip it",
		IsConfirm: true,
	}
	_, err := prompt.Run()
	if err != nil {
		return fmt.Errorf("canceled")
	}
	return nil
}
