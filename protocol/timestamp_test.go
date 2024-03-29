/*
Copyright (c) Facebook, Inc. and its affiliates.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package protocol

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Testing conversion so if Packet structure changes we notice
func Test_byteToTime(t *testing.T) {
	timeb := []byte{63, 155, 21, 96, 0, 0, 0, 0, 52, 156, 191, 42, 0, 0, 0, 0}
	res, err := byteToTime(timeb)
	require.Nil(t, err)

	require.Equal(t, int64(1612028735717200436), res.UnixNano())
}

func Test_ReadTXtimestamp(t *testing.T) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	require.Nil(t, err)
	defer conn.Close()

	txts, attempts, err := ReadTXtimestamp(conn)
	require.Equal(t, time.Time{}, txts)
	require.Equal(t, maxTXTS, attempts)
	require.Equal(t, fmt.Errorf("no TX timestamp found after %d tries", maxTXTS), err)

	err = EnableSWTimestampsSocket(conn)
	require.Nil(t, err)

	addr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 12345}
	_, err = conn.WriteTo([]byte{}, addr)
	require.Nil(t, err)
	txts, attempts, err = ReadTXtimestamp(conn)

	require.NotEqual(t, time.Time{}, txts)
	require.Equal(t, 1, attempts)
	require.Nil(t, err)

}

func Test_scmDataToTime(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		want    int64
		wantErr bool
	}{
		{
			name: "hardware timestamp",
			data: []byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				63, 155, 21, 96, 0, 0, 0, 0, 52, 156, 191, 42, 0, 0, 0, 0,
			},
			want:    1612028735717200436,
			wantErr: false,
		},
		{
			name: "software timestamp",
			data: []byte{
				63, 155, 21, 96, 0, 0, 0, 0, 52, 156, 191, 42, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			},
			want:    1612028735717200436,
			wantErr: false,
		},
		{
			name: "zero timestamp",
			data: []byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := scmDataToTime(tt.data)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.Nil(t, err)
				require.Equal(t, tt.want, res.UnixNano())
			}
		})
	}
}
