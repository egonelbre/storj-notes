package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/egonelbre/storj-notes/notes"

	"storj.io/uplink"
)

func main() {
	// handle Ctrl-C
	ctx, cancel := ContextWithSignal(context.Background(), os.Interrupt)
	defer cancel()

	// flag handling
	passphrase := flag.String("passphrase", os.Getenv("NOTES_PASSPHRASE"), "passphrase for data")
	apikey := flag.String("apikey", os.Getenv("NOTES_APIKEY"), "apikey for the satellite")
	satellite := flag.String("satellite", os.Getenv("NOTES_SATELLITE"), "satellite address for notes")

	accessString := flag.String("access", os.Getenv("NOTES_ACCESS"), "access grant to storj network")

	bucket := flag.String("bucket", os.Getenv("NOTES_BUCKET"), "bucket name")
	if *bucket == "" {
		*bucket = "notes"
	}

	flag.Parse()

	hasPassword := *passphrase != "" && *apikey != "" && *satellite != ""
	hasAccess := *accessString != ""

	if !hasPassword && !hasAccess {
		fmt.Fprintln(os.Stderr, "Authentication information not set:")
		fmt.Fprintln(os.Stderr, "* --passphrase, --apikey, --satellite")
		fmt.Fprintln(os.Stderr, "* --access")
		fmt.Fprintln(os.Stderr)

		flag.Usage()
		os.Exit(1)
	}

	command := flag.Arg(0)
	identifier := flag.Arg(1)
	if command == "" {
		fmt.Fprintln(os.Stderr, "Command not set `list`, `get`, `set`, `delete`")
		fmt.Fprintln(os.Stderr)

		flag.Usage()
		os.Exit(1)
	}

	// open the access
	var access *uplink.Access
	var err error
	if hasPassword {
		access, err = uplink.RequestAccessWithPassphrase(ctx, *satellite, *apikey, *passphrase)
	} else {
		access, err = uplink.ParseAccess(*accessString)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to load access grant: %v\n", err)
		os.Exit(1)
	}

	// start the service
	service, err := notes.Open(ctx, access, *bucket)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to open notes service: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := service.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to close service: %v\n", err)
		}
	}()

	// handle different commands
	switch command {
	case "get":
		note, err := service.Get(ctx, identifier)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to get note %q: %v\n", identifier, err)
			os.Exit(1)
		}
		fmt.Println(note)
	case "set":
		value := flag.Arg(2)
		err := service.Set(ctx, identifier, value)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to set note %q: %v\n", identifier, err)
			os.Exit(1)
		}
	case "list":
		identifiers, err := service.List(ctx, identifier)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to set note %q: %v\n", identifier, err)
			os.Exit(1)
		}
		for _, identifier := range identifiers {
			fmt.Println(identifier)
		}
	case "delete":
		err := service.Delete(ctx, identifier)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to delete note %q: %v\n", identifier, err)
			os.Exit(1)
		}
	default:
		os.Exit(1)
	}
}

// ContextWithSignal creates a context that will be cancelled on the signals.
func ContextWithSignal(ctx context.Context, signals ...os.Signal) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, signals...)

	go func() {
		select {
		case <-ch:
			cancel()
		case <-ctx.Done():
		}
	}()

	return ctx, func() {
		cancel()
		signal.Stop(ch)
	}
}
