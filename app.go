package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/urfave/cli/v2"
)

type app struct {
	cli *cli.App
	destination
	flag
}

type destination struct {
	completion string
	domain     cli.StringSlice
	list       string
	output     string
	timeout    time.Duration
	insecure   bool
	noTimeInfo bool
}

type flag struct {
	completion *cli.StringFlag
	domain     *cli.StringSliceFlag
	list       *cli.PathFlag
	output     *cli.StringFlag
	timeout    *cli.DurationFlag
	insecure   *cli.BoolFlag
	noTimeInfo *cli.BoolFlag
}

func newApp() *app {
	a := app{}
	a.flag.completion = &cli.StringFlag{
		Name:        "completion",
		Aliases:     []string{"c"},
		Usage:       fmt.Sprintf("completion scripts: %s", pipeJoin(shells)),
		Destination: &a.destination.completion,
	}
	a.flag.domain = &cli.StringSliceFlag{
		Name:        "domain",
		Aliases:     []string{"d"},
		Usage:       "domain:port separated by commas",
		Destination: &a.destination.domain,
	}
	a.flag.list = &cli.PathFlag{
		Name:        "list",
		Aliases:     []string{"l"},
		Usage:       "path to newline-delimited list of domains",
		Destination: &a.destination.list,
	}
	a.flag.output = &cli.StringFlag{
		Name:        "output",
		Aliases:     []string{"o"},
		Usage:       fmt.Sprintf("output format: %s", pipeJoin(formats)),
		Destination: &a.destination.output,
		Value:       formatJSON.String(),
	}
	a.flag.timeout = &cli.DurationFlag{
		Name:        "timeout",
		Aliases:     []string{"t"},
		Usage:       "network timeout: ns|us|ms|s|m|h",
		Destination: &a.destination.timeout,
		Value:       5 * time.Second,
	}
	a.flag.insecure = &cli.BoolFlag{
		Name:        "insecure",
		Aliases:     []string{"i"},
		Usage:       "skip verification of the cert chain and host name",
		Destination: &a.destination.insecure,
		Value:       false,
	}
	a.flag.noTimeInfo = &cli.BoolFlag{
		Name:        "no-timeinfo",
		Aliases:     []string{"n"},
		Usage:       "hide fields related to the current time in table output",
		Destination: &a.destination.noTimeInfo,
		Value:       false,
	}
	a.cli = &cli.App{
		Name:                 Name,
		Usage:                "TLS cert checker CLI",
		Version:              Version,
		Description:          "CLI application for checking TLS certificate information",
		HideHelpCommand:      true,
		EnableBashCompletion: true,
		Before:               a.doValidate,
		Action:               a.doAction,
		Flags:                []cli.Flag{a.flag.completion, a.flag.domain, a.flag.list, a.flag.output, a.flag.timeout, a.flag.insecure, a.flag.noTimeInfo},
	}
	return &a
}

func (a *app) run(ctx context.Context) error {
	return a.cli.RunContext(ctx, os.Args)
}

func (a *app) doAction(c *cli.Context) error {
	if c.IsSet(a.flag.completion.Name) {
		return comp(a.destination.completion)
	}
	var domains []string
	var err error
	switch {
	case c.IsSet(a.flag.domain.Name):
		domains = a.destination.domain.Value()
	case c.IsSet(a.flag.list.Name):
		domains, err = fromList(a.destination.list)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("cannot parse flags: cannot receive domain names from %s or %s", a.flag.domain.Name, a.flag.list.Name)
	}
	list, err := getCertList(c.Context, domains, a.destination.timeout, a.destination.insecure)
	if err != nil {
		return err
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].DomainName < list[j].DomainName
	})
	out, err := out(list, a.destination.output, a.destination.noTimeInfo)
	if err != nil {
		return err
	}
	fmt.Println(out)
	return nil
}

func (a *app) doValidate(c *cli.Context) error {
	if err := checkProvided(c); err != nil {
		return err
	}
	if err := checkSingle(c, a.flag.completion.Name, []string{a.flag.domain.Name, a.flag.list.Name, a.flag.output.Name, a.flag.timeout.Name, a.flag.insecure.Name, a.flag.noTimeInfo.Name}); err != nil {
		return err
	}
	if err := checkValidPair(c, a.flag.domain.Name, a.flag.list.Name); err != nil {
		return err
	}
	if c.Bool(a.flag.insecure.Name) {
		if err := insecureConfirm(); err != nil {
			return err
		}
	}
	return nil
}

func checkProvided(c *cli.Context) error {
	if c.Args().Len() == 0 && c.NumFlags() == 0 {
		return fmt.Errorf("cannot parse command line flags: no flag provided")
	}
	return nil
}

func checkSingle(c *cli.Context, target string, flags []string) error {
	if c.IsSet(target) {
		for _, flag := range flags {
			if c.IsSet(flag) {
				return fmt.Errorf("cannot parse command line flags: %s is not available when other flags are set", target)
			}
		}
	}
	return nil
}

func checkValidPair(c *cli.Context, a string, b string) error {
	if c.IsSet(a) && c.IsSet(b) {
		return fmt.Errorf("cannot parse command line flags: cannot be used together %s and %s", a, b)
	}
	return nil
}

func insecureConfirm() error {
	ni, _ := strconv.ParseBool(os.Getenv(strings.ToUpper(Name) + "_NON_INTERACTIVE"))
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

func pipeJoin(s []string) string {
	return strings.Join(s, "|")
}
