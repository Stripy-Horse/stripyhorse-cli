package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	stripyhorse "github.com/Stripy-Horse/stripyhorse-go"
)

func cmdPrinters(cfg *Config, args []string) error {
	if len(args) == 0 {
		return errors.New("usage: stripyhorse printers <list|create|delete> …")
	}
	switch args[0] {
	case "list", "ls":
		return printersList(cfg, args[1:])
	case "create", "new":
		return printersCreate(cfg, args[1:])
	case "delete", "rm":
		return printersDelete(cfg, args[1:])
	default:
		return fmt.Errorf("unknown printers subcommand %q (list|create|delete)", args[0])
	}
}

func printersList(cfg *Config, _ []string) error {
	client, ctx, err := cfg.apiClient()
	if err != nil {
		return err
	}
	res, _, err := client.SimulatorAPI.ListPrinters(ctx).Execute()
	if err != nil {
		return apiError(err)
	}
	printers := res.GetPrinters()
	if len(printers) == 0 {
		fmt.Fprintln(os.Stderr, "no printers yet — create one with `stripyhorse printers create`")
		return nil
	}
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tNAME\tMODE\tTCP\tANON")
	for _, p := range printers {
		tcp := p.GetTcp()
		anon := ""
		if p.GetAnonymize() {
			anon = "yes"
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s:%d\t%s\n",
			p.GetId(), p.GetName(), p.GetMode(), tcp.GetHost(), tcp.GetPort(), anon)
	}
	return tw.Flush()
}

func printersCreate(cfg *Config, args []string) error {
	fs := flag.NewFlagSet("printers create", flag.ExitOnError)
	name := fs.String("name", "my-printer", "printer name")
	preset := fs.String("size", "4x6", "label size preset, e.g. 4x6")
	mode := fs.String("mode", "ephemeral", "ephemeral or persistent")
	anon := fs.Bool("anonymize", false, "mask PII + strip graphics from captured jobs")
	fs.Parse(args)

	client, ctx, err := cfg.apiClient()
	if err != nil {
		return err
	}
	body := stripyhorse.NewCreatePrinterInputBody(*name)
	body.SetPreset(*preset)
	body.SetMode(*mode)
	body.SetAnonymize(*anon)

	p, _, err := client.SimulatorAPI.CreatePrinter(ctx).CreatePrinterInputBody(*body).Execute()
	if err != nil {
		return apiError(err)
	}

	// Cache the ingest URL — it's only returned here, and `print` needs it.
	if p.GetIngestUrl() != "" {
		cfg.Ingest[p.GetId()] = p.GetIngestUrl()
		_ = cfg.save()
	}
	tcp := p.GetTcp()
	fmt.Printf("Created %s (%s)\n", p.GetId(), p.GetName())
	fmt.Printf("  TCP:    %s:%d\n", tcp.GetHost(), tcp.GetPort())
	fmt.Printf("  Ingest: %s\n", p.GetIngestUrl())
	fmt.Printf("  Print:  stripyhorse print --printer %s <file.zpl>\n", p.GetId())
	return nil
}

func printersDelete(cfg *Config, args []string) error {
	fs := flag.NewFlagSet("printers delete", flag.ExitOnError)
	fs.Parse(args)
	if fs.NArg() < 1 {
		return errors.New("usage: stripyhorse printers delete <printer-id>")
	}
	id := fs.Arg(0)
	client, ctx, err := cfg.apiClient()
	if err != nil {
		return err
	}
	if _, err := client.SimulatorAPI.DeletePrinter(ctx, id).Execute(); err != nil {
		return apiError(err)
	}
	delete(cfg.Ingest, id)
	_ = cfg.save()
	fmt.Fprintf(os.Stderr, "deleted %s\n", id)
	return nil
}
