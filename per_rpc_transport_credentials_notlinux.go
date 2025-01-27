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

// +build !linux,!windows

package grpcfd

import (
	"os"
	"sync"
)

func (w *wrapPerRPCCredentials) SendFilename(filename string) <-chan error {
	out := make(chan error, 1)
	// Note: this will fail in most cases for 'unopenable' files (like unix file sockets).  See use of O_PATH in per_rpc_transport_credentials_linux.go for
	// the trick that makes this work in Linux
	file, err := os.Open(filename) // #nosec
	if err != nil {
		out <- err
		close(out)
		return out
	}
	var wg sync.WaitGroup
	w.executor.AsyncExec(func() {
		w.senderFuncs = append(w.senderFuncs, func(sender FDSender) {
			wg.Add(1)

			go func() {
				if sender != nil {
					defer wg.Done()
					joinErrChs(sender.SendFile(file), out)
				} else {
					wg.Done()
					wg.Wait()
					_ = file.Close()
					close(out)
				}
			}()
		})
	})
	return out
}
