package main

import (
	"bytes"
	"testing"

	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/golang/mock/gomock"
)

func TestDNSServers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	project, tables, fs, pf, tf := mockFactories(ctrl)

	servers := []belvedere.DNSServer{
		{
			Hostname: "one",
		},
	}

	project.EXPECT().
		DNSServers(gomock.Any()).
		Return(servers, nil)

	tables.EXPECT().
		Print(servers)

	cmd := newRootCmd("test").ToCobra(pf, tf, fs)
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetArgs([]string{
		"dns-servers",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}
