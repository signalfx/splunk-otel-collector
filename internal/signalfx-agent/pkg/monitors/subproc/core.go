// Package subproc holds the logic for managing monitors that run in a
// subprocess.
package subproc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

// RuntimeConfig for subprocs
type RuntimeConfig struct {
	Binary string
	Args   []string
	// Envvars in the form "key=value".
	Env []string
}

// RuntimeCustomizable can be implemented by runners that use MonitorCore
// to provide extra config about the subprocess runtime to use.
type RuntimeCustomizable interface {
	RuntimeConfig() RuntimeConfig
}

// MonitorCore is the adapter to the subprocess monitor runner.  It communiates
// with the subprocess using named pipes.  Each general type of subprocess
// monitor (e.g. Datadog, collectd, Java, etc.) should get its own generic monitor
// struct that uses this adapter by embedding it.
//
// This will run a single, dedicated subprocess that actually runs the
// monitoring code.  Getting data/metrics/events out of the subprocess is the
// responsibility of modules that embed this MonitorCore, hence there are no
// predefined "datapoint" message types.
type MonitorCore struct {
	ctx     context.Context
	cancel  func()
	handler MessageHandler

	logger log.FieldLogger

	// Conditional signal that the goroutine that sends does the configuration
	// request sets when configure has been completed.  configResult will hold
	// the result of that configure call.
	configCond   sync.Cond
	configResult error

	// Flag that should be set atomically to tell the goroutine that manages
	// the subprocess whether the process is supposed to be alive or not.
	shutdownCalled int32
}

// New returns a new uninitialized monitor core
func New() *MonitorCore {
	ctx, cancel := context.WithCancel(context.Background())

	return &MonitorCore{
		logger:     log.StandardLogger(),
		ctx:        ctx,
		cancel:     cancel,
		configCond: sync.Cond{L: &sync.Mutex{}},
	}

}

// Logger returns the logger that should be used
func (mc *MonitorCore) Logger() log.FieldLogger {
	return mc.logger
}

// run the subprocess and block until it returns.  Messages from stderr will be
// logged as error logs in the agent.
func (mc *MonitorCore) run(runtimeConf RuntimeConfig, stdin io.Reader, stdout io.Writer) error {
	mc.logger.Debugf("Subprocess command: %s %v (env: %v)", runtimeConf.Binary, runtimeConf.Args, runtimeConf.Env)

	cmd := exec.CommandContext(mc.ctx, runtimeConf.Binary, runtimeConf.Args...)
	cmd.SysProcAttr = procAttrs()
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Env = runtimeConf.Env

	// Stderr is just the normal output from the subprocess that isn't
	// specially encoded
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	mc.logger = mc.logger.WithFields(log.Fields{
		"runnerPID": cmd.Process.Pid,
	})
	mc.logger.Info("Started subprocess runner")

	go func() {
		scanner := utils.ChunkScanner(stderr)
		for scanner.Scan() {
			mc.logger.Error(scanner.Text())
		}
	}()

	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}

// run the subprocess, restarting it if it stops while this monitor is still
// active.
func (mc *MonitorCore) runWithRestart(runtimeConf RuntimeConfig, handler MessageHandler, configBytes []byte) {
	for {
		messages, stdin, stdout, err := makePipes()
		if err != nil {
			mc.logger.WithError(err).Error("Couldn't create pipes for subprocess monitor")
			return
		}

		go func() {
			mc.configCond.L.Lock()
			mc.configResult = mc.doConfigure(messages, configBytes)
			mc.configCond.L.Unlock()
			// Tell the initial Configure method call that the subproc is done
			// configuring.
			mc.configCond.Broadcast()

			if mc.configResult != nil {
				mc.Shutdown()
				mc.logger.WithError(mc.configResult).Error("Could not configure subprocess monitor")
				return
			}

			handler.ProcessMessages(mc.ctx, messages)
		}()

		err = mc.run(runtimeConf, stdin, stdout)
		mc.configCond.Broadcast()

		stdin.Close()
		stdout.Close()
		messages.Close()

		if err != nil {
			mc.logger.WithError(err).Error("Subprocess monitor runner shutdown with error")
		}
		if mc.ShutdownCalled() {
			return
		}
		mc.logger.Error("Restarting subprocess runner")

		time.Sleep(2 * time.Second)
	}
}

func makePipes() (*messageReadWriter, io.ReadCloser, io.WriteCloser, error) {
	stdinReader, stdinWriter, err := os.Pipe()
	// If this errors, things are really wrong with the system
	if err != nil {
		return nil, nil, nil, err
	}

	stdoutReader, stdoutWriter, err := os.Pipe()
	// If this errors, things are really wrong with the system
	if err != nil {
		return nil, nil, nil, err
	}

	return &messageReadWriter{
		Reader: stdoutReader,
		Writer: stdinWriter,
	}, stdinReader, stdoutWriter, nil
}

// MessageHandler is what monitors that use MonitorCore should provide to deal
// with messages that come in after a successful Configure call to the
// subprocess.  MontiorCore.ConfigureInSubproc handles configuring the
// subprocess monitor and collecting the result.
type MessageHandler interface {
	ProcessMessages(context.Context, MessageReceiver)
	HandleLogMessage(io.Reader) error
}

// ConfigureInSubproc sends the given config to the subproc and returns whether
// configuration was successful.  This method should only be called once for
// the lifetime of the monitor.  The returned MessageReceiver can be used to
// get datapoints/events out of the subprocess, the exact format of the data is
// left up to the users of this core.
func (mc *MonitorCore) ConfigureInSubproc(config config.MonitorCustomConfig,
	runtimeConfig *RuntimeConfig, handler MessageHandler) error {
	if mc.handler != nil {
		panic("ConfigureInSubproc should only be called once")
	}

	mc.handler = handler
	mc.logger = mc.logger.WithFields(log.Fields{
		"monitorID":   config.MonitorConfigCore().MonitorID,
		"monitorType": config.MonitorConfigCore().Type,
	})

	jsonBytes, err := json.Marshal(config)
	if err != nil {
		return err
	}

	mc.configCond.L.Lock()
	defer mc.configCond.L.Unlock()

	go mc.runWithRestart(*runtimeConfig, handler, jsonBytes)
	mc.configCond.Wait()

	return mc.configResult
}

func (mc *MonitorCore) doConfigure(messages *messageReadWriter, jsonBytes []byte) error {
	if err := messages.SendMessage(MessageTypeConfigure, jsonBytes); err != nil {
		return err
	}

	result, err := mc.waitForConfigure(messages)
	if err != nil {
		return err
	}

	if result.Error != nil {
		return errors.New(*result.Error)
	}

	return nil
}

func (mc *MonitorCore) waitForConfigure(messages MessageReceiver) (*configResult, error) {
	for {
		msgType, payloadReader, err := messages.RecvMessage()
		if err != nil {
			return nil, err
		}

		content, err := ioutil.ReadAll(payloadReader)
		if err != nil {
			mc.logger.WithError(err).Error("Could not read message from subprocess monitor")
		}
		payloadReader = bytes.NewBuffer(content)

		switch msgType {
		case MessageTypeConfigureResult:
			var result configResult
			if err := json.NewDecoder(payloadReader).Decode(&result); err != nil {
				return nil, err
			}
			return &result, nil
		case MessageTypeLog:
			if err := mc.handler.HandleLogMessage(payloadReader); err != nil {
				mc.logger.WithError(err).Error("Could not read log message from subprocess monitor")
			}
		default:
			return nil, fmt.Errorf("got unexpected message code %d from subprocess monitor", msgType)
		}
	}
}

// ShutdownCalled returns true if the Shutdown method has been called.
func (mc *MonitorCore) ShutdownCalled() bool {
	return atomic.LoadInt32(&mc.shutdownCalled) > 0
}

// Shutdown the whole Runner child process, not just individual monitors
func (mc *MonitorCore) Shutdown() {
	atomic.StoreInt32(&mc.shutdownCalled, 1)

	mc.cancel()
}
