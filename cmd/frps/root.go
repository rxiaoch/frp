package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"

	"github.com/fatedier/frp/g"
	"github.com/fatedier/frp/models/config"
	"github.com/fatedier/frp/server"
	"github.com/fatedier/frp/utils/log"
	"github.com/fatedier/frp/utils/version"
)

const (
	CfgFileTypeIni = iota
	CfgFileTypeCmd
)

var (
	cfgFile     string
	showVersion bool

	bindAddr          string
	bindPort          int
	bindUdpPort       int
	kcpBindPort       int
	proxyBindAddr     string
	vhostHttpPort     int
	vhostHttpsPort    int
	dashboardAddr     string
	dashboardPort     int
	dashboardUser     string
	dashboardPwd      string
	assetsDir         string
	logFile           string
	logWay            string
	logLevel          string
	logMaxDays        int64
	token             string
	authTimeout       int64
	subDomainHost     string
	tcpMux            bool
	allowPorts        string
	maxPoolCount      int64
	maxPortsPerClient int64
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "", "c", "", "config file of frps")
	rootCmd.PersistentFlags().BoolVarP(&showVersion, "version", "v", false, "version of frpc")

	rootCmd.PersistentFlags().StringVarP(&bindAddr, "bind_addr", "", "0.0.0.0", "bind address")
	rootCmd.PersistentFlags().IntVarP(&bindPort, "bind_port", "p", 7000, "bind port")
	rootCmd.PersistentFlags().IntVarP(&bindUdpPort, "bind_udp_port", "", 0, "bind udp port")
	rootCmd.PersistentFlags().IntVarP(&kcpBindPort, "kcp_bind_port", "", 0, "kcp bind udp port")
	rootCmd.PersistentFlags().StringVarP(&proxyBindAddr, "proxy_bind_addr", "", "0.0.0.0", "proxy bind address")
	rootCmd.PersistentFlags().IntVarP(&vhostHttpPort, "vhost_http_port", "", 0, "vhost http port")
	rootCmd.PersistentFlags().IntVarP(&vhostHttpsPort, "vhost_https_port", "", 0, "vhost https port")
	rootCmd.PersistentFlags().StringVarP(&dashboardAddr, "dashboard_addr", "", "0.0.0.0", "dasboard address")
	rootCmd.PersistentFlags().IntVarP(&dashboardPort, "dashboard_port", "", 0, "dashboard port")
	rootCmd.PersistentFlags().StringVarP(&dashboardUser, "dashboard_user", "", "admin", "dashboard user")
	rootCmd.PersistentFlags().StringVarP(&dashboardPwd, "dashboard_pwd", "", "admin", "dashboard password")
	rootCmd.PersistentFlags().StringVarP(&logFile, "log_file", "", "console", "log file")
	rootCmd.PersistentFlags().StringVarP(&logWay, "log_way", "", "console", "log way")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log_level", "", "info", "log level")
	rootCmd.PersistentFlags().Int64VarP(&logMaxDays, "log_max_days", "", 3, "log_max_days")
	rootCmd.PersistentFlags().StringVarP(&token, "token", "", "", "auth token")
	rootCmd.PersistentFlags().Int64VarP(&authTimeout, "auth_timeout", "", 900, "auth timeout")
	rootCmd.PersistentFlags().StringVarP(&subDomainHost, "subdomain_host", "", "", "subdomain host")
	rootCmd.PersistentFlags().Int64VarP(&maxPortsPerClient, "max_ports_per_client", "", 0, "max ports per client")
}

var rootCmd = &cobra.Command{
	Use:   "frps",
	Short: "frps is the server of frp (https://github.com/fatedier/frp)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if showVersion {
			fmt.Println(version.Full())
			return nil
		}

		if cfgFile != "" {
			parseServerCommonCfg(CfgFileTypeIni, cfgFile)
		} else {
			parseServerCommonCfg(CfgFileTypeCmd, "")
		}

		err := runServer()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func parseServerCommonCfg(fileType int, filePath string) (err error) {
	if fileType == CfgFileTypeIni {
		err = parseServerCommonCfgFromIni(filePath)
	} else if fileType == CfgFileTypeCmd {
		err = parseServerCommonCfgFromCmd()
	}
	if err != nil {
		return
	}

	g.GlbServerCfg.CfgFile = cfgFile

	err = g.GlbServerCfg.ServerCommonConf.Check()
	if err != nil {
		return
	}

	config.InitServerCfg(&g.GlbServerCfg.ServerCommonConf)
	return
}

func parseServerCommonCfgFromIni(filePath string) (err error) {
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	content := string(b)

	cfg, err := config.UnmarshalServerConfFromIni(&g.GlbServerCfg.ServerCommonConf, content)
	if err != nil {
		return err
	}
	g.GlbServerCfg.ServerCommonConf = *cfg
	return
}

func parseServerCommonCfgFromCmd() (err error) {
	g.GlbServerCfg.BindAddr = bindAddr
	g.GlbServerCfg.BindPort = bindPort
	g.GlbServerCfg.BindUdpPort = bindUdpPort
	g.GlbServerCfg.KcpBindPort = kcpBindPort
	g.GlbServerCfg.ProxyBindAddr = proxyBindAddr
	g.GlbServerCfg.VhostHttpPort = vhostHttpPort
	g.GlbServerCfg.VhostHttpsPort = vhostHttpsPort
	g.GlbServerCfg.DashboardAddr = dashboardAddr
	g.GlbServerCfg.DashboardPort = dashboardPort
	g.GlbServerCfg.DashboardUser = dashboardUser
	g.GlbServerCfg.DashboardPwd = dashboardPwd
	g.GlbServerCfg.LogFile = logFile
	g.GlbServerCfg.LogWay = logWay
	g.GlbServerCfg.LogLevel = logLevel
	g.GlbServerCfg.LogMaxDays = logMaxDays
	g.GlbServerCfg.Token = token
	g.GlbServerCfg.AuthTimeout = authTimeout
	g.GlbServerCfg.SubDomainHost = subDomainHost
	g.GlbServerCfg.MaxPortsPerClient = maxPortsPerClient
	return
}

func runServer() (err error) {
	log.InitLog(g.GlbServerCfg.LogWay, g.GlbServerCfg.LogFile, g.GlbServerCfg.LogLevel,
		g.GlbServerCfg.LogMaxDays)
	svr, err := server.NewService()
	if err != nil {
		return err
	}
	log.Info("Start frps success")
	server.ServerService = svr
	svr.Run()
	return
}
