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

	domain := &cli.StringSliceFlag{
		Name:    "domain",
		Aliases: []string{"d"},
		Usage:   "domain:port separated by commas",
	}

	file := &cli.StringFlag{
		Name:    "file",
		Aliases: []string{"f"},
		Usage:   "path to newline-delimited list of domains",
	}

	output := &cli.StringFlag{
		Name:    "output",
		Aliases: []string{"o"},
		Usage:   "set output type",
		Value:   "json",
		Sources: cli.EnvVars(label + "_OUTPUT_TYPE"),
		//	Value:   tlc3.OutputTypeCompressedText.String(),
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

	noTimeInfo := &cli.BoolFlag{
		Name:    "no-timeinfo",
		Aliases: []string{"n"},
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

		if err := checkValidPair(cmd, domain.Name, file.Name); err != nil {
			return nil, err
		}

		if cmd.Bool(insecure.Name) {
			if err := insecureConfirm(); err != nil {
				return nil, err
			}
		}

		return ctx, nil
	}

	action := func(ctx context.Context, cmd *cli.Command) error {
		var domains []string
		var err error
		if cmd.IsSet(domain.Name) {
			domains = cmd.StringSlice(domain.Name)
		}
		if cmd.IsSet(file.Name) {
			domains, err = tlc3.FromList(cmd.String(file.Name))
			if err != nil {
				return err
			}
		}
		if len(domains) == 0 {
			return errors.New("cannot receive domain names")
		}
		tz := cmd.String(timeZone.Name)
		loc, err := time.LoadLocation(tz)
		if err != nil {
			return fmt.Errorf("cannot load timezone %q", tz)
		}
		logger.Info("getting certificate information...")
		infos, err := tlc3.GetCertList(ctx, domains, cmd.Duration(timeout.Name), cmd.Bool(insecure.Name), loc)
		if err != nil {
			return err
		}
		slices.SortFunc(infos, func(a, b *tlc3.CertInfo) int {
			return cmp.Compare(a.DomainName, b.DomainName)
		})
		if err := tlc3.Out(infos, w, cmd.String(output.Name), cmd.Bool(noTimeInfo.Name)); err != nil {
			return err
		}
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
		Flags:                 []cli.Flag{loglevel, domain, file, output, timeout, insecure, noTimeInfo, timeZone},
	}
}

func checkValidPair(cmd *cli.Command, a string, b string) error {
	if cmd.IsSet(a) && cmd.IsSet(b) {
		return fmt.Errorf("cannot be used together %s and %s", a, b)
	}
	return nil
}

func insecureConfirm() error {
	ni, _ := strconv.ParseBool(os.Getenv(label + "_NON_INTERACTIVE"))
	if ni {
		return nil
	}
	prompt := promptui.Prompt{
		Label:     "[WARNING] insecure flag skips verification of the certificate chain and hostname. skip it",
		IsConfirm: true,
	}
	_, err := prompt.Run()
	if err != nil {
		return err
	}
	return nil
}
