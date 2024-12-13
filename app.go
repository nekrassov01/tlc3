package main

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"time"

	"github.com/charmbracelet/log"
	"github.com/manifoldco/promptui"
	"github.com/urfave/cli/v2"
)

const (
	appName       = "tlc3"
	canonicalName = "TLC3"
)

type app struct {
	*cli.App
	completion *cli.StringFlag
	loglevel   *cli.StringFlag
	domain     *cli.StringSliceFlag
	file       *cli.PathFlag
	output     *cli.StringFlag
	timeout    *cli.DurationFlag
	insecure   *cli.BoolFlag
	noTimeInfo *cli.BoolFlag
	timeZone   *cli.StringFlag
}

func CLI(ctx context.Context) {
	logger := log.New(os.Stderr).WithPrefix(canonicalName)
	log.SetDefault(logger)
	app := newApp(os.Stdout)
	if err := app.RunContext(ctx, os.Args); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

func newApp(w io.Writer) *app {
	a := app{}
	a.completion = &cli.StringFlag{
		Name:    "completion",
		Aliases: []string{"c"},
		Usage:   fmt.Sprintf("completion scripts: %s", pipeJoin(shells)),
	}
	a.loglevel = &cli.StringFlag{
		Name:    "log-level",
		Aliases: []string{"l"},
		Usage:   "log levels: debug|info|warn|error",
		Value:   log.InfoLevel.String(),
		EnvVars: []string{canonicalName + "_LOGLEVEL"},
	}
	a.domain = &cli.StringSliceFlag{
		Name:    "domain",
		Aliases: []string{"d"},
		Usage:   "domain:port separated by commas",
	}
	a.file = &cli.PathFlag{
		Name:    "file",
		Aliases: []string{"f"},
		Usage:   "path to newline-delimited list of domains",
	}
	a.output = &cli.StringFlag{
		Name:    "output",
		Aliases: []string{"o"},
		Usage:   fmt.Sprintf("output format: %s", pipeJoin(formats)),
		Value:   formatJSON.String(),
		EnvVars: []string{canonicalName + "_OUTPUT"},
	}
	a.timeout = &cli.DurationFlag{
		Name:    "timeout",
		Aliases: []string{"t"},
		Usage:   "network timeout: ns|us|ms|s|m|h",
		Value:   5 * time.Second,
		EnvVars: []string{canonicalName + "_TIMEOUT"},
	}
	a.insecure = &cli.BoolFlag{
		Name:    "insecure",
		Aliases: []string{"i"},
		Usage:   "skip verification of the cert chain and host name",
		Value:   false,
	}
	a.noTimeInfo = &cli.BoolFlag{
		Name:    "no-timeinfo",
		Aliases: []string{"n"},
		Usage:   "hide fields related to the current time in table output",
		Value:   false,
	}
	a.timeZone = &cli.StringFlag{
		Name:    "timezone",
		Aliases: []string{"z"},
		Usage:   "time zone for datetime fields",
		Value:   "Local",
		EnvVars: []string{canonicalName + "_TIMEZONE"},
	}
	a.App = &cli.App{
		Name:                 appName,
		Usage:                "TLS cert checker CLI",
		Version:              Version,
		Writer:               w,
		Description:          "CLI application for checking TLS certificate information",
		HideHelpCommand:      true,
		EnableBashCompletion: true,
		Before:               a.before,
		Action:               a.action,
		Flags:                []cli.Flag{a.completion, a.loglevel, a.domain, a.file, a.output, a.timeout, a.insecure, a.noTimeInfo, a.timeZone},
	}
	return &a
}

func (a *app) before(c *cli.Context) error {
	var (
		target = a.completion.Name
		flags  = []string{
			a.loglevel.Name,
			a.domain.Name,
			a.file.Name,
			a.output.Name,
			a.timeout.Name,
			a.insecure.Name,
			a.noTimeInfo.Name,
		}
	)
	if err := checkSingle(c, target, flags); err != nil {
		return err
	}
	if err := checkValidPair(c, a.domain.Name, a.file.Name); err != nil {
		return err
	}
	if c.Bool(a.insecure.Name) {
		if err := insecureConfirm(); err != nil {
			return err
		}
	}
	level, err := log.ParseLevel(c.String(a.loglevel.Name))
	if err != nil {
		return err
	}
	log.SetLevel(level)
	return nil
}

func (a *app) action(c *cli.Context) error {
	if c.NumFlags() == 0 {
		return cli.ShowAppHelp(c)
	}
	if c.IsSet(a.completion.Name) {
		return comp(a.Writer, c.String(a.completion.Name))
	}
	var domains []string
	var err error
	if c.IsSet(a.domain.Name) {
		domains = c.StringSlice(a.domain.Name)
	}
	if c.IsSet(a.file.Name) {
		domains, err = fromList(c.Path(a.file.Name))
		if err != nil {
			return err
		}
	}
	if len(domains) == 0 {
		return errors.New("cannot receive domain names")
	}
	tz := c.String(a.timeZone.Name)
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return fmt.Errorf("cannot load timezone %q", tz)
	}
	log.Info("getting certificate information...")
	infos, err := getCertList(c.Context, domains, c.Duration(a.timeout.Name), c.Bool(a.insecure.Name), loc)
	if err != nil {
		return err
	}
	slices.SortFunc(infos, func(a, b *certInfo) int {
		return cmp.Compare(a.DomainName, b.DomainName)
	})
	if err := out(infos, a.Writer, c.String(a.output.Name), c.Bool(a.noTimeInfo.Name)); err != nil {
		return err
	}
	log.Info("completed")
	return nil
}

func checkSingle(c *cli.Context, target string, flags []string) error {
	if !c.IsSet(target) {
		return nil
	}
	for _, flag := range flags {
		if c.IsSet(flag) {
			return fmt.Errorf("%s: not available if other flags are set", target)
		}
	}
	return nil
}

func checkValidPair(c *cli.Context, a string, b string) error {
	if c.IsSet(a) && c.IsSet(b) {
		return fmt.Errorf("cannot be used together %s and %s", a, b)
	}
	return nil
}

func insecureConfirm() error {
	ni, _ := strconv.ParseBool(os.Getenv(canonicalName + "_NON_INTERACTIVE"))
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
