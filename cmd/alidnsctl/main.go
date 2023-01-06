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
				Name:     ArgsAccessKeyID,
				Usage:    "accessKeyID for auth",
				Required: true,
				EnvVars:  []string{"ACCESS_KEY_ID"},
			},
			&cli.StringFlag{
				Name:     ArgsAccessKeySecret,
				Usage:    "accessKeySecret for auth",
				Required: true,
				EnvVars:  []string{"ACCESS_KEY_SECRET"},
			},
			&cli.StringFlag{
				Name:    "output",
				Usage:   "output format",
				Aliases: []string{"o"},
			},
		},
		Commands: []*cli.Command{
			CompltionCommand,
			{
				Name: "domains",
				Subcommands: []*cli.Command{
					{
						Name: "list",
						Action: ctlfunc(func(ctx *cli.Context, ctl *alidnsctl.AliDNSCtl) (any, error) {
							return ctl.ListDomains(ctx.Context)
						}),
					}, {
						Name:      "get",
						UsageText: `alidns domains get <domain>`,
						Aliases:   []string{"info"},
						Action: ctlfunc(func(ctx *cli.Context, ctl *alidnsctl.AliDNSCtl) (any, error) {
							return ctl.GetDomain(ctx.Context, ctx.Args().First())
						}),
					},
				},
			},
			{
				Name: "records",
				Subcommands: []*cli.Command{
					{
						Name: "list",
						Flags: []cli.Flag{
							&cli.StringFlag{Name: "domain", Required: true},
						},
						Action: ctlfunc(func(ctx *cli.Context, ctl *alidnsctl.AliDNSCtl) (any, error) {
							return ctl.ListRecords(ctx.Context, ctx.String("domain"), "")
						}),
					},
					{
						Name:      "get",
						Before:    RequiredArgs,
						UsageText: `alidns records get <record id>`,
						Action: ctlfunc(func(ctx *cli.Context, ctl *alidnsctl.AliDNSCtl) (any, error) {
							return ctl.GetRecord(ctx.Context, ctx.Args().First())
						}),
					},
					{
						Name: "set",
						UsageText: `
alidns records set --domain example.com --type=A --rr=www --value=127.0.0.1
alidns records set --domain example.com --type=TXT --rr=www --value=5DBVzPVw9HPF
alidns records set --domain example.com --type=CNAME --rr=@ --value=www.example.com
alidns records set --domain example.com --type=AAAA --rr=*.6 --value=fe80::98c4:e4ff:fe94:cc67
`,
						Flags: []cli.Flag{
							&cli.StringFlag{Name: "domain", Required: true},
							&cli.StringFlag{Name: "type", Required: true},
							&cli.StringFlag{Name: "value", Required: true},
							&cli.StringFlag{Name: "rr", Required: true},
						},
						Action: ctlfunc(func(ctx *cli.Context, ctl *alidnsctl.AliDNSCtl) (any, error) {
							result, err := ctl.SetRecord(ctx.Context,
								ctx.String("domain"),
								ctx.String("type"),
								ctx.String("rr"),
								ctx.String("value"),
							)
							if err != nil {
								return nil, err
							}
							return ctl.GetRecord(ctx.Context, *result.RecordId)
						}),
					},
					{
						Name:      "enable",
						Before:    RequiredArgs,
						UsageText: `alidns records enable <record id>`,
						Action: ctlfunc(func(ctx *cli.Context, ctl *alidnsctl.AliDNSCtl) (any, error) {
							return ctl.SetRecordStatus(ctx.Context, ctx.Args().First(), false)
						}),
					},
					{
						Name:      "disable",
						Before:    RequiredArgs,
						UsageText: `alidns records disable <record id>`,
						Action: ctlfunc(func(ctx *cli.Context, ctl *alidnsctl.AliDNSCtl) (any, error) {
							return ctl.SetRecordStatus(ctx.Context, ctx.Args().First(), true)
						}),
					},
					{
						Name:      "remove",
						Aliases:   []string{"delete"},
						UsageText: `alidns records remove <record id>`,
						Before:    RequiredArgs,
						Action: ctlfunc(func(ctx *cli.Context, ctl *alidnsctl.AliDNSCtl) (any, error) {
							return ctl.DeleteRecord(ctx.Context, ctx.Args().First())
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
