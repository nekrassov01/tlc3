package main

import (
	"cmp"
	"context"
	"crypto/tls"
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
	logger = log.NewLogger(log.NewCLIHandler(io.Discard))
)

func newCmd(w, ew io.Writer) *cli.Command {
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
		Value:   tlc3.OutputTypeText.String(),
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

	tlsVersion := &cli.StringFlag{
		Name:    "tls-version",
		Aliases: []string{"m"},
		Usage:   "tls minimum version to use",
		Value:   tlc3.DefaultTLSVersionString,
		Sources: cli.EnvVars(label + "_TLS_VERSION"),
	}

	before := func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
		// parse log level
		var level slog.Level
		if err := level.UnmarshalText([]byte(cmd.String(loglevel.Name))); err != nil {
			level = slog.LevelInfo
		}

		// create logger for application
		s := log.Style2()
		s.Caller.Fullpath = true
		logger = log.NewLogger(log.NewCLIHandler(ew,
			log.WithLabel("TLC3"),
			log.WithTime(true),
			log.WithLevel(level),
			log.WithCaller(level <= slog.LevelDebug),
			log.WithStyle(s),
		))

		// check flags combinations
		if cmd.IsSet(addr.Name) && cmd.IsSet(file.Name) {
			return nil, fmt.Errorf("cannot be used together %q and %q", addr.Name, file.Name)
		}

		// confirm insecure flag
		if cmd.Bool(insecure.Name) {
			if err := confirmInsecure(); err != nil {
				return nil, err
			}
		}

		// confirm old TLS version
		version, err := tlc3.ParseTLSVersion(cmd.String(tlsVersion.Name))
		if err != nil {
			return nil, err
		}
		if version < tls.VersionTLS12 {
			if err := confirmTLSVersion(); err != nil {
				return nil, err
			}
		}
		cmd.Metadata[tlsVersion.Name] = version

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
		version := cmd.Metadata[tlsVersion.Name].(uint16)
		infos, err := tlc3.GetCerts(ctx, addresses, cmd.Duration(timeout.Name), cmd.Bool(insecure.Name), loc, version)
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
		Version:               tlc3.Version(),
		Usage:                 "TLS cert checker CLI",
		Description:           "CLI application for checking TLS certificate informations",
		HideHelpCommand:       true,
		EnableShellCompletion: true,
		Writer:                w,
		ErrWriter:             ew,
		Before:                before,
		Action:                action,
		Flags:                 []cli.Flag{loglevel, addr, file, output, timeout, insecure, static, timeZone, tlsVersion},
		Metadata:              map[string]any{},
	}
}

// confirmInsecure prompts the user to confirm the use of the insecure flag.
func confirmInsecure() error {
	return confirm("[WARNING] insecure flag skips verification of the certificate chain and hostname. skip it")
}

// confirmTLSVersion prompts the user to confirm the use of an old TLS version.
func confirmTLSVersion() error {
	return confirm("[WARNING] We recommend using TLS version 1.2 or higher. Do you wish to proceed despite the risk")
}

// confirm prompts the user to confirm the action.
func confirm(msg string) error {
	ci, _ := strconv.ParseBool(os.Getenv(label + "_NON_INTERACTIVE"))
	if ci {
		return nil
	}
	prompt := promptui.Prompt{
		Label:     msg,
		IsConfirm: true,
	}
	_, err := prompt.Run()
	if err != nil {
		return fmt.Errorf("canceled")
	}
	return nil
}
