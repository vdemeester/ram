package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dnephin/filewatcher/files"
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/vdemeester/ram/runner"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var (
	defaultExcludes = []string{
		"**/*.swp",
		"**/*.swx",
		"vendor/",
		".git",
		".ignore",
		".gitignore",
		"**/*file",
		"*.lock",
	}
)

type options struct {
	verbose     bool
	quiet       bool
	exclude     []string
	dirs        []string
	depth       int
	command     []string
	idleTimeout time.Duration
	events      eventOpt
}

func setupFlags(args []string) *options {
	flags := pflag.NewFlagSet(args[0], pflag.ContinueOnError)
	flags.ParseErrorsWhitelist = pflag.ParseErrorsWhitelist{UnknownFlags: true}
	opts := options{}
	flags.BoolVarP(&opts.verbose, "verbose", "v", false, "Verbose")
	flags.BoolVarP(&opts.quiet, "quiet", "q", false, "Quiet")
	flags.StringSliceVarP(&opts.exclude, "exclude", "x", nil, "Exclude file patterns")
	flags.StringSliceVarP(&opts.dirs, "directory", "d", []string{"."}, "Directories to watch")
	flags.IntVarP(&opts.depth, "depth", "L", 5, "Descend only level directories deep")
	flags.DurationVar(&opts.idleTimeout, "idle-timeout", 10*time.Minute,
		"Exit after idle timeout")
	flags.VarP(&opts.events, "event", "e",
		"events to watch (create,write,remove,rename,chmod)")

	flags.SetInterspersed(false)
	flags.Usage = func() {
		out := os.Stderr
		fmt.Fprintf(out, "Usage:\n  %s [OPTIONS] COMMAND ARGS... \n\n", os.Args[0])
		fmt.Fprint(out, "Options:\n")
		flags.PrintDefaults()
	}
	if err := flags.Parse(args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
		os.Exit(2)
	}
	opts.command = flags.Args()
	return &opts
}

func main() {
	opts := setupFlags(os.Args)
	setupLogging(opts)

	run(opts)
}

func run(opts *options) {
	excludes := append(defaultExcludes, opts.exclude...)
	excludeList, err := files.NewExcludeList(excludes)
	if err != nil {
		log.Fatalf("Error creating exclude list: %s", err)
	}

	dirs := files.WalkDirectories(opts.dirs, opts.depth, excludeList)
	watcher, err := buildWatcher(dirs)
	if err != nil {
		log.Fatalf("Error setting up watcher: %s", err)
	}
	defer watcher.Close()

	log.Debugf("Handling events: %s", opts.events.Value())
	command := append([]string{"go", "test"}, cleanCommand(opts.command)...)
	log.Infof("Run %s", strings.Join(command, " "))
	handler, cleanup := runner.NewRunner(excludeList, opts.events.Value(), command)
	defer cleanup()
	watchOpts := runner.WatchOptions{
		IdleTimeout: opts.idleTimeout,
		Runner:      handler,
	}
	if err = runner.Watch(watcher, watchOpts); err != nil {
		log.Fatalf("Error during watch: %s", err)
	}
}

func cleanCommand(command []string) []string {
	c := []string{}
	for _, v := range command {
		if v == "--" {
			continue
		}
		c = append(c, v)
	}
	return append(c, "./${dir}")
}

func setupLogging(opts *options) {
	formatter := new(prefixed.TextFormatter)
	formatter.DisableTimestamp = true
	log.SetFormatter(formatter)
	if opts.verbose {
		log.SetLevel(log.DebugLevel)
	}
	if opts.quiet {
		log.SetLevel(log.WarnLevel)
	}
}

func buildWatcher(dirs []string) (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	log.Infof("Watching directories: %s", strings.Join(dirs, ", "))
	for _, dir := range dirs {
		log.Debugf("Adding new watch: %s", dir)
		if err = watcher.Add(dir); err != nil {
			return nil, err
		}
	}
	return watcher, nil
}
