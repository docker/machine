package bugsnag

import (
	"log"
	"os"
	"testing"

	"github.com/juju/loggo"
)

func TestNotifyReleaseStages(t *testing.T) {

	var testCases = []struct {
		stage      string
		configured []string
		notify     bool
		msg        string
	}{
		{
			stage:  "production",
			notify: true,
			msg:    "Should notify in all release stages by default",
		},
		{
			stage:      "production",
			configured: []string{"development", "production"},
			notify:     true,
			msg:        "Failed to notify in configured release stage",
		},
		{
			stage:      "staging",
			configured: []string{"development", "production"},
			notify:     false,
			msg:        "Failed to prevent notification in excluded release stage",
		},
	}

	for _, testCase := range testCases {
		Configure(Configuration{ReleaseStage: testCase.stage, NotifyReleaseStages: testCase.configured})

		if Config.notifyInReleaseStage() != testCase.notify {
			t.Error(testCase.msg)
		}
	}
}

func TestProjectPackages(t *testing.T) {
	Configure(Configuration{ProjectPackages: []string{"main", "github.com/ConradIrwin/*"}})
	if !Config.isProjectPackage("main") {
		t.Error("literal project package doesn't work")
	}
	if !Config.isProjectPackage("github.com/ConradIrwin/foo") {
		t.Error("wildcard project package doesn't work")
	}
	if Config.isProjectPackage("runtime") {
		t.Error("wrong packges being marked in project")
	}
	if Config.isProjectPackage("github.com/ConradIrwin/foo/bar") {
		t.Error("wrong packges being marked in project")
	}

}

type LoggoWrapper struct {
	loggo.Logger
}

func (lw *LoggoWrapper) Printf(format string, v ...interface{}) {
	lw.Logger.Warningf(format, v...)
}

func TestConfiguringCustomLogger(t *testing.T) {

	l1 := log.New(os.Stdout, "", log.Lshortfile)

	l2 := &LoggoWrapper{loggo.GetLogger("test")}

	var testCases = []struct {
		config Configuration
		notify bool
		msg    string
	}{
		{
			config: Configuration{ReleaseStage: "production", NotifyReleaseStages: []string{"development", "production"}, Logger: l1},
			notify: true,
			msg:    "Failed to assign log.Logger",
		},
		{
			config: Configuration{ReleaseStage: "production", NotifyReleaseStages: []string{"development", "production"}, Logger: l2},
			notify: true,
			msg:    "Failed to assign LoggoWrapper",
		},
	}

	for _, testCase := range testCases {
		Configure(testCase.config)

		// call printf just to illustrate it is present as the compiler does most of the hard work
		testCase.config.Logger.Printf("hello %s", "bugsnag")

	}
}
