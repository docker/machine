package bugsnag

import (
	"errors"

	"github.com/Sirupsen/logrus"
	"github.com/bugsnag/bugsnag-go"
	bugsnag_errors "github.com/bugsnag/bugsnag-go/errors"
)

type Hook struct{}

// ErrBugsnagUnconfigured is returned if NewBugsnagHook is called before
// bugsnag.Configure. Bugsnag must be configured before the hook.
var ErrBugsnagUnconfigured = errors.New("bugsnag must be configured before installing this logrus hook")

// ErrBugsnagSendFailed indicates that the hook failed to submit an error to
// bugsnag. The error was successfully generated, but `bugsnag.Notify()`
// failed.
type ErrBugsnagSendFailed struct {
	err error
}

func (e ErrBugsnagSendFailed) Error() string {
	return "failed to send error to Bugsnag: " + e.err.Error()
}

// NewBugsnagHook initializes a logrus hook which sends exceptions to an
// exception-tracking service compatible with the Bugsnag API. Before using
// this hook, you must call bugsnag.Configure(). The returned object should be
// registered with a log via `AddHook()`
//
// Entries that trigger an Error, Fatal or Panic should now include an "error"
// field to send to Bugsnag.
func NewBugsnagHook() (*Hook, error) {
	if bugsnag.Config.APIKey == "" {
		return nil, ErrBugsnagUnconfigured
	}
	return &Hook{}, nil
}

// skipStackFrames skips logrus stack frames before logging to Bugsnag.
const skipStackFrames = 4

var accumulator = []*logrus.Entry{}

// Fire forwards an error to Bugsnag. Given a logrus.Entry, it extracts the
// "error" field (or the Message if the error isn't present) and sends it off.
func (hook *Hook) Fire(entry *logrus.Entry) error {
	accumulator = append(accumulator, entry)
	if entry.Level == logrus.DebugLevel || entry.Level == logrus.InfoLevel {
		return nil
	}

	var notifyErr error
	err, ok := entry.Data["error"].(error)
	if ok {
		notifyErr = err
	} else {
		notifyErr = errors.New(entry.Message)
	}

	var tabs = bugsnag.MetaData{}

	for k, v := range context {
		for kk, vv := range v {
			tabs.Add(k, kk, vv)
		}
	}

	tabs["data"] = make(map[string]interface{})

	for key, value := range entry.Data {
		hook.Add("data", key, value)
	}
	/*for _, ent := range accumulator{
		t := fmt.Sprint(ent.Time)
		l := ent.Level
		m := ent.Message
		msg := fmt.Sprintf("%s: %s", l, m)
		hook.Add("log", t, msg)
	}*/

	errWithStack := bugsnag_errors.New(notifyErr, skipStackFrames)
	bugsnagErr := bugsnag.Notify(errWithStack, tabs)
	if bugsnagErr != nil {
		return ErrBugsnagSendFailed{bugsnagErr}
	}

	return nil
}

var context = map[string]map[string]interface{}{}

func (hook *Hook) Add(tab string, key string, value interface{}) {
	if context[tab] == nil {
		context[tab] = make(map[string]interface{})
	}
	context[tab][key] = value
}

// Levels enumerates the log levels on which the error should be forwarded to
// bugsnag: everything at or above the "Error" level.
func (hook *Hook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.DebugLevel,
		logrus.InfoLevel,
		logrus.WarnLevel,
		logrus.ErrorLevel,
		logrus.FatalLevel,
		logrus.PanicLevel,
	}
}
