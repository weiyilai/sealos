package main

import (
	"flag"
	"net"
	"net/http"

	"github.com/cesanta/glog"
	"github.com/labring/sealos/service/hub/server"
)

type RestartableServer struct {
	configFile string
	authServer *server.AuthServer
	hs         *http.Server
}

func (rs *RestartableServer) Serve(c *server.Config) {
	// NewAuthServer will start a go routine to reset reqLimiter
	as, err := server.NewAuthServer(c)
	if err != nil {
		glog.Exitf("Failed to create auth server: %s", err)
	}
	hs := &http.Server{
		Addr:    c.Server.ListenAddress,
		Handler: as,
	}

	rs.authServer, rs.hs = as, hs
	var listener net.Listener
	listener, err = net.Listen("tcp", c.Server.ListenAddress)
	if err != nil {
		glog.Fatal(err.Error())
	}

	glog.Infof("Serving on %s", c.Server.ListenAddress)
	if err := hs.Serve(listener); err != nil {
		glog.Fatal(err.Error())
	}
}

func main() {
	flag.Parse()
	glog.CopyStandardLogTo("INFO")

	cf := flag.Arg(0)
	if cf == "" {
		glog.Exitf("Config file not specified")
	}
	config, err := server.LoadConfig(cf)
	if err != nil {
		glog.Exitf("Failed to load config: %s", err)
	}
	rs := RestartableServer{
		configFile: cf,
	}
	rs.Serve(config)
}