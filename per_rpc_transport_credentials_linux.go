//  Copyright (c) 2020 Cisco and/or its affiliates.
//
//  SPDX-License-Identifier: Apache-2.0
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

// +build linux

package grpcfd

import (
	"os"
	"sync"

	"golang.org/x/sys/unix"
)

func (w *wrapPerRPCCredentials) SendFilename(filename string) <-chan error {
	out := make(chan error, 1)
	file, err := os.OpenFile(filename, unix.O_PATH, 0) // #nosec
	if err != nil {
		out <- err
		close(out)
		return out
	}
	var wg sync.WaitGroup
	w.executor.AsyncExec(func() {
		w.senderFuncs = append(w.senderFuncs, func(sender FDSender) {
			go func() {
				if sender != nil {
					wg.Add(1)
					defer wg.Done()
					joinErrChs(sender.SendFile(file), out)
				} else {
					wg.Wait()
					_ = file.Close()
					close(out)
				}
			}()
		})
	})
	return out
}
