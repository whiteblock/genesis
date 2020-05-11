/**
 * Copyright 2019 Whiteblock Inc. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package auxillary

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/whiteblock/genesis/pkg/config"
	"github.com/whiteblock/genesis/pkg/entity"
	"github.com/whiteblock/genesis/pkg/usecase"

	"github.com/innodv/errors/await"
	"github.com/sirupsen/logrus"
	"github.com/whiteblock/definition/command"
	"golang.org/x/sync/semaphore"
)

// Executor handles the  processing of mutliple commands
type Executor interface {
	ExecuteCommands(cmds []command.Command) entity.Result
	Prepare(inst *command.Instructions) error
}

type executor struct {
	usecase usecase.DockerUseCase
	conf    config.Execution
	log     logrus.Ext1FieldLogger
}

// ErrDockerConnFailed is the error for when the docker daemon is unreachable
var ErrDockerConnFailed = entity.NewFatalResult("could not connect to docker")

// NewExecutor creates a new DeliveryHandler which uses the given usecase for
// executing the extracted command
func NewExecutor(
	conf config.Execution,
	usecase usecase.DockerUseCase,
	log logrus.Ext1FieldLogger) Executor {
	return &executor{usecase: usecase, conf: conf, log: log}
}

func (exec executor) Prepare(inst *command.Instructions) error {
	dir := filepath.Join("/tmp", inst.ID)
	_, err := os.Lstat(dir)
	if err == nil { //already exists
		return nil
	}
	err = os.Mkdir(dir, 0755)
	if err != nil {
		return err
	}
	exec.log.WithFields(logrus.Fields{
		"testID": inst.ID,
		"dir":    dir,
	}).Info("creating the tls auth files")
	errChan := make(chan error)
	go func() {
		errChan <- ioutil.WriteFile(filepath.Join(dir, "ca.cert"), inst.Auth.CACertPEM(), 0644)
	}()

	go func() {
		errChan <- ioutil.WriteFile(filepath.Join(dir, "client.cert"), inst.Auth.ClientCertPEM(), 0644)
	}()

	go func() {
		errChan <- ioutil.WriteFile(filepath.Join(dir, "client.key"), inst.Auth.ClientPKPEM(), 0644)
	}()
	return await.AwaitErrors(errChan, 3)
}

func (exec executor) ExecuteCommands(cmds []command.Command) entity.Result {
	resultChan := make(chan entity.Result, len(cmds))
	sem := semaphore.NewWeighted(exec.conf.LimitPerTest)
	ctx, cancelFn := context.WithTimeout(context.Background(), exec.conf.TimeLimit)
	defer cancelFn()
	for _, cmd := range cmds {
		go func(cmd command.Command) {
			for i := 0; i < exec.conf.ConnectionRetries; i++ {
				err := sem.Acquire(ctx, 1)
				if err != nil {
					exec.log.WithFields(logrus.Fields{
						"error": err,
						"cmd":   cmd,
					}).Debug("received a cancelation signal")
					resultChan <- entity.NewSuccessResult() // successfully killed
					return
				}

				res := exec.usecase.Run(ctx, cmd)
				sem.Release(1)
				if !res.IsSuccess() && strings.Contains(res.Error.Error(), "connect to the Docker daemon") {
					exec.log.WithFields(logrus.Fields{
						"result":  res,
						"time":    exec.conf.RetryDelay,
						"attempt": i,
					}).Info("connection to docker failed, retrying")
					time.Sleep(exec.conf.RetryDelay)
					continue
				}
				resultChan <- res.InjectMeta(map[string]interface{}{
					"command": cmd,
					"attempt": i,
				})
				return
			}
			resultChan <- ErrDockerConnFailed.InjectMeta(
				map[string]interface{}{
					"command": cmd,
				})
		}(cmd)
	}
	var err error
	isTrap := false
	failed := []string{}
	var propagatedResult entity.Result
	for range cmds {
		result := <-resultChan
		entry := exec.log.WithField("result", result)

		entry.Trace("finished processing a command")
		if result.IsDelayed() {
			entry.Debug("result contains a delay, propogating it")
			propagatedResult = result
		} else if result.IsFatal() {
			entry.Error("a command had a fatal error")
			cancelFn()
			propagatedResult = result
		} else if !result.IsSuccess() {
			failed = append(failed, result.Meta["command"].(command.Command).ID)
			entry.Warn("a command failed to execute")
			if err != nil {
				err = fmt.Errorf("%v;%v", err, result.Error.Error())
			} else {
				err = result.Error
			}

		} else if result.IsTrap() {
			entry.Info("a command raised a trap")
			isTrap = true
		}
	}
	if propagatedResult.IsFatal() || propagatedResult.IsDelayed() { // was there a fatal error? If so, just return that
		return propagatedResult
	}
	if err != nil {
		return entity.NewErrorResult(err).InjectMeta(map[string]interface{}{
			"failed": failed,
		})
	}
	if isTrap {
		return entity.NewTrapResult()
	}
	return entity.NewSuccessResult()
}
