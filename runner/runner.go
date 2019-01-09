package runner

import (
	"bufio"
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/dnephin/filewatcher/files"
	"github.com/dnephin/filewatcher/ui"
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
)

// Runner executes commands when an included file is modified
type Runner struct {
	excludes *files.ExcludeList
	command  []string
	events   chan fsnotify.Event
	eventOp  fsnotify.Op
}

// NewRunner creates a new Runner
func NewRunner(
	excludes *files.ExcludeList,
	eventOp fsnotify.Op,
	command []string,
) (*Runner, func()) {
	events := make(chan fsnotify.Event)
	return &Runner{
		excludes: excludes,
		command:  command,
		events:   events,
		eventOp:  eventOp,
	}, func() { close(events) }
}

func (runner *Runner) start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-runner.events:
			// FIXME: I'm not sure how this empty event gets created
			if event.Name == "" && event.Op == 0 {
				return
			}
			runner.run(event)
		}
	}
}

// HandleEvent checks runs the command if the event was a Write event
func (runner *Runner) HandleEvent(event fsnotify.Event) {
	if !runner.shouldHandle(event) {
		return
	}

	// Send the event to an unbuffered channel so that on events floods only
	// one event is run, and the rest are dropped.
	select {
	case runner.events <- event:
	default:
		log.Debugf("Events queued, skipping: %s", event.Name)
	}
}

func (runner *Runner) run(event fsnotify.Event) {
	start := time.Now()
	command := runner.buildCommand(event.Name)
	ui.PrintStart(command)

	err := run(command, event.Name)
	ui.PrintEnd(time.Since(start), event.Name, err)
}

func (runner *Runner) shouldHandle(event fsnotify.Event) bool {
	if event.Op&runner.eventOp == 0 {
		log.Debugf("Skipping excluded event: %s (%v)", event.Op, event.Op&runner.eventOp)
		return false
	}

	filename := event.Name
	if runner.excludes.IsMatch(filename) {
		log.Debugf("Skipping excluded file: %s", filename)
		return false
	}

	return true
}

func (runner *Runner) buildCommand(filename string) []string {
	mapping := func(key string) string {
		switch key {
		case "filepath":
			return filename
		case "dir":
			return path.Dir(filename)
		case "relative_dir":
			return addDotSlash(filepath.Dir(filename))
		}
		return key
	}

	output := []string{}
	buildtag, err := detectBuildTags(filename)
	if err != nil {
		log.Warn(err)
	}
	if buildtag != "" {
		output = append(output, "-tags="+buildtag)
	}
	switch {
	case strings.HasSuffix(filename, "_test.go"):
		fs := token.NewFileSet()
		f, err := os.Open(filename)
		if err != nil {
			log.Warn(err)
		}
		ff, err := parser.ParseFile(fs, filename, f, parser.AllErrors)
		if err != nil {
			log.Warn(err)
		}
		v := &testVisitor{}
		ast.Walk(v, ff)
		for _, arg := range runner.command {
			output = append(output, os.Expand(arg, mapping))
		}
		if len(v.tests) != 0 {
			output = append(output, "-test.run", "^"+strings.Join(v.tests, "|")+"$")
		}
	default:
		for _, arg := range runner.command {
			output = append(output, os.Expand(arg, mapping))
		}
	}
	return output
}

func detectBuildTags(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	sc := bufio.NewScanner(f)
	if sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "// +build") {
			buildtags := strings.Fields(strings.TrimPrefix(line, "// +build "))
			if len(buildtags) != 1 {
				log.Warnf("detected multiple build tags, not supported : %s", line)
				return "", nil
			}
			return buildtags[0], nil
		}
	}
	return "", nil
}

type testVisitor struct {
	tests []string
}

func (v *testVisitor) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		return nil
	}
	switch d := n.(type) {
	case *ast.FuncDecl:
		if strings.HasPrefix(d.Name.Name, "Test") {
			v.tests = append(v.tests, d.Name.Name)
		}
	}
	return v
}

func run(command []string, filename string) error {
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(),
		"TEST_DIRECTORY="+addDotSlash(filepath.Dir(filename)),
		"TEST_FILENAME="+addDotSlash(filename))
	return cmd.Run()
}

func addDotSlash(filename string) string {
	return "." + string(filepath.Separator) + filename
}
