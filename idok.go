package main

import (
	"flag"
	"fmt"
	"github.com/metal3d/idok/asserver"
	"github.com/metal3d/idok/tunnel"
	"github.com/metal3d/idok/utils"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

const (
	VERSION = "v1-alpha2"
)

func main() {

	log.SetFlags(0)

	// flags
	xbmcaddr := flag.String("target", "", "xbmc/kodi ip (raspbmc address, ip or hostname)")
	username := flag.String("login", "", "jsonrpc login (configured in xbmc settings)")
	password := flag.String("password", "", "jsonrpc password (configured in xbmc settings)")
	viassh := flag.Bool("ssh", false, "use SSH Tunnelling (need ssh user and password)")
	nossh := flag.Bool("nossh", false, "force to not use SSH tunnel - usefull to override configuration file")
	port := flag.Int("port", 8080, "local port (ignored if you use ssh option)")
	sshuser := flag.String("sshuser", "pi", "ssh login")
	sshpassword := flag.String("sshpass", "", "ssh password")
	sshport := flag.Int("sshport", 22, "target ssh port")
	version := flag.Bool("version", false, fmt.Sprintf("Print the current version (%s)", VERSION))
	xbmcport := flag.Int("targetport", 80, "XBMC/Kodi jsonrpc port")
	stdin := flag.Bool("stdin", false, "read file from stdin to stream")
	confexample := flag.Bool("conf-example", false, "print a configuration file example to STDOUT")

	flag.Usage = utils.Usage

	flag.Parse()

	// print the current version
	if *version {
		fmt.Println(VERSION)
		fmt.Println("Compiled for", runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	}

	// If user asks to prints configuration file example, print it and exit
	if *confexample {
		utils.PrintExampleConfig()
		os.Exit(0)
	}

	// Set new configuration from options
	conf := &utils.Config{
		Target:      *xbmcaddr,
		Targetport:  *xbmcport,
		Localport:   *port,
		User:        *username,
		Password:    *password,
		Sshuser:     *sshuser,
		Sshpassword: *sshpassword,
		Sshport:     *sshport,
		Ssh:         *viassh,
	}

	// check if conf file exists and override options
	if filename, found := utils.CheckLocalConfigFiles(); found {
		utils.LoadLocalConfig(filename, conf)
	}

	if conf.Target == "" {
		fmt.Println("Error: you must provide the xbmc server address")
		//flag.Usage()
		os.Exit(1)
	}

	utils.SetTarget(conf)

	var dir, file string

	// we don't use stdin, so we should check if scheme is file, youtube or other...
	if !*stdin {
		if len(flag.Args()) < 1 {
			fmt.Println("Error: you must provide a stream or a file")
			//flag.Usage()
			os.Exit(2)
		}

		if youtube, vid := utils.IsYoutubeURL(flag.Arg(0)); youtube {
			log.Println("Sending Youtube stream URL...")
			utils.PlayYoutube(vid)
			os.Exit(0)
		}

		if ok, local := utils.IsOtherScheme(flag.Arg(0)); ok {
			log.Println("Sending stream URL...")
			utils.SendBasicStream(flag.Arg(0), local)
			os.Exit(0)
		}

		// find the good path
		toserve := flag.Arg(0)
		dir = "."
		toserve, _ = filepath.Abs(toserve)
		file = filepath.Base(toserve)
		dir = filepath.Dir(toserve)

	}

	if conf.Ssh && !*nossh {
		config := tunnel.NewConfig(*sshuser, *sshpassword)
		// serve ssh tunnel !
		if !*stdin {
			tunnel.SshHTTPForward(config, file, dir)
		} else {
			tunnel.SshForwardStdin(config)
		}
	} else {
		// serve local port !
		if !*stdin {
			asserver.HttpServe(file, dir, *port)
		} else {
			asserver.TCPServeStdin(*port)
		}
	}
}
