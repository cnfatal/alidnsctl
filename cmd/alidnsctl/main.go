package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/cnfatal/alidnsctl"
	"github.com/urfave/cli/v2"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill, syscall.SIGTERM)
	defer cancel()
	if err := NewAliDNSCtlCommand().RunContext(ctx, os.Args); err != nil {
		fmt.Printf("%s\n", err.Error())
		os.Exit(1)
	}
}

const (
	ArgsAccessKeyID     = "accessKeyID"
	ArgsAccessKeySecret = "accessKeySecret"
)

func NewAliDNSCtlCommand() *cli.App {
	app := &cli.App{
		Name:                 "alidnsctl",
		Usage:                "alidns command line tool",
		EnableBashCompletion: true,
		Version:              alidnsctl.Version(),
		Suggest:              true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    ArgsAccessKeyID,
				Usage:   "accessKeyID for auth",
				Value:   "",
				EnvVars: []string{"ACCESS_KEY_ID"},
			},
			&cli.StringFlag{
				Name:    ArgsAccessKeySecret,
				Usage:   "accessKeySecret for auth",
				Value:   "",
				EnvVars: []string{"ACCESS_KEY_SECRET"},
			},
			&cli.StringFlag{
				Name:        "output",
				Aliases:     []string{"o"},
				Usage:       "output format",
				DefaultText: "json",
			},
		},
		Commands: []*cli.Command{
			CompltionCommand,
			{
				Name:   "get",
				Usage:  "query dns records",
				Before: RequiredArgs,
				UsageText: `
# list all types of records
alidnsctl get example.com

# list all A records
alidnsctl get --type A api.example.com

# list all AAAA records
alidnsctl get --type AAAA 6.example.com

# list CNAME records
alidnsctl get --type CNAME www.example.com`,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name: "type", Aliases: []string{"t"},
					},
				},
				Action: ctlfunc(func(ctx *cli.Context, ctl *alidnsctl.AliDNSCtl) (any, error) {
					fulldomain, typ := ctx.Args().First(), ctx.String("type")
					return ctl.ListRecords(ctx.Context, fulldomain, typ)
				}),
			},
			{
				Name:    "set",
				Usage:   "set dns records",
				Aliases: []string{"apply"},
				UsageText: `
apply will set records to the values provided in args;
the records not in the list will be removed.

# apply two A records. (will remove records which value not in this args)
alidnsctl set www.example.com 127.0.0.1 127.0.0.2

# apply two AAAA records.
alidnsctl set --type=AAAA www.example.com fe80::98c4:e4ff:fe94:cc67 fe80::98c4:e4ff:fe94:cc68

alidnsctl set --type=CNAME www.example.com foo.com

alidnsctl set --type=TXT example.com 5DBVzPVw9HPF`,
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "type", Aliases: []string{"t"}, Value: "A"},
				},
				Before: RequiredArgs,
				Action: ctlfunc(func(ctx *cli.Context, ctl *alidnsctl.AliDNSCtl) (any, error) {
					fulldomain, typ := ctx.Args().First(), ctx.String("type")
					if ctx.Args().Len() > 1 {
						values := ctx.Args().Slice()[1:]
						if err := ctl.SetRecord(ctx.Context, fulldomain, typ, values); err != nil {
							return nil, err
						}
					}
					return ctl.ListRecords(ctx.Context, fulldomain, typ)
				}),
			},
			{
				Name:    "del",
				Usage:   "remove dns records",
				Aliases: []string{"remove"},
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "all"},
					&cli.StringFlag{Name: "type", Aliases: []string{"t"}, Value: "A"},
				},
				UsageText: `
# remove A records 127.0.0.1 127.0.0.2 from www.example.com
alidnsctl del www.example.com 127.0.0.1 127.0.0.2

# remove ipv6 records
alidnsctl del --type AAAA www.example.com fe80::98c4:e4ff:fe94:cc67 fe80::98c4:e4ff:fe94:cc68

# remove cname records
alidnsctl del --type CNAME www.example.com foo.com

# remove all A records on www.example.com if no type set.
alidnsctl del --all www.example.com

# remove all AAAA records on www.example.com.
alidnsctl del --all --type AAAA www.example.com

alidnsctl del --type=TXT example.com 5DBVzPVw9HPF`,
				Before: RequiredArgs,
				Action: ctlfunc(func(ctx *cli.Context, ctl *alidnsctl.AliDNSCtl) (any, error) {
					fulldomain, typ := ctx.Args().First(), ctx.String("type")
					if ctx.Bool("all") {
						if _, err := ctl.DeleteRecordBatch(ctx.Context, fulldomain, typ); err != nil {
							return nil, err
						}
					}
					values := ctx.Args().Slice()[1:]
					if err := ctl.DeleteRecordFromValues(ctx.Context, fulldomain, typ, values); err != nil {
						return nil, err
					}
					return ctl.ListRecords(ctx.Context, fulldomain, typ)
				}),
			},
			{
				Name:  "domains",
				Usage: "list available domains",
				Subcommands: []*cli.Command{
					{
						Name:  "list",
						Usage: "list all domains",
						Action: ctlfunc(func(ctx *cli.Context, ctl *alidnsctl.AliDNSCtl) (any, error) {
							return ctl.ListDomains(ctx.Context)
						}),
					},
					{
						Name:      "info",
						UsageText: `alidnsctl domains info <domain>`,
						Action: ctlfunc(func(ctx *cli.Context, ctl *alidnsctl.AliDNSCtl) (any, error) {
							return ctl.GetDomain(ctx.Context, ctx.Args().First())
						}),
					},
				},
			},
		},
	}
	return app
}

func ctlfunc(fun func(ctx *cli.Context, ctl *alidnsctl.AliDNSCtl) (any, error)) func(ctx *cli.Context) error {
	return func(ctx *cli.Context) error {
		ctl, err := alidnsctl.New(alidnsctl.Options{
			AccesskeyID:     ctx.String(ArgsAccessKeyID),
			AccessKeySecret: ctx.String(ArgsAccessKeySecret),
		})
		if err != nil {
			return err
		}
		data, err := fun(ctx, ctl)
		if err != nil {
			return err
		}
		if data != nil {
			return printResult(data, ctx.String("output"))
		} else {
			return nil
		}
	}
}

func printResult(data any, format string) error {
	switch format {
	case "json", "":
		fallthrough
	default:
		raw, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return err
		}
		fmt.Print(string(raw))
	}
	return nil
}

func RequiredArgs(cCtx *cli.Context) error {
	if cCtx.Args().Len() == 0 {
		cCtx.Command.Command(cli.HelpFlag.Names()[0]).Action(cCtx)
		return fmt.Errorf("")
	}
	return nil
}
