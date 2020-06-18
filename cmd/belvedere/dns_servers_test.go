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

	project, output, pf, of := mockFactories(ctrl)

	servers := []belvedere.DNSServer{
		{
			Hostname: "one",
		},
	}

	project.EXPECT().
		DNSServers(gomock.Any()).
		Return(servers, nil)

	output.EXPECT().
		Print(servers)

	cmd := newRootCmd("test").ToCobra(pf, of)
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetErr(bytes.NewBuffer(nil))
	cmd.SetArgs([]string{
		"dns-servers",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}
