package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/pivotalservices/cfbackup"
	cfopsplugin "github.com/pivotalservices/cfops/plugin/cfopsplugin"
	"github.com/pivotalservices/gtils/command"
	"github.com/xchapter7x/lo"
)

var (
	//NewRemoteExecuter -
	NewRemoteExecuter = command.NewRemoteExecutor
)

func main() {
	cfopsplugin.Start(NewNFSPlugin())
}

//GetMeta - method to provide metadata
func (s *NFSPlugin) GetMeta() (meta cfopsplugin.Meta) {
	meta = s.Meta
	return
}

//Setup - on setup method
func (s *NFSPlugin) Setup(pcf cfopsplugin.PivotalCF) (err error) {
	s.PivotalCF = pcf
	s.InstallationSettings = pcf.GetInstallationSettings()
	s.parsePluginArgs()
	return
}
func (s *NFSPlugin) parsePluginArgs() {
	const (
		productName = "--productName"
		jobName     = "--jobName"
	)
	args := strings.Split(s.PivotalCF.GetHostDetails().PluginArgs, " ")
	if len(args) != 4 {
		panic("Must specify --productName and --jobName as plugin args")
	}
	if args[0] == productName {
		s.ProductName = args[1]
	}
	if args[0] == jobName {
		s.JobName = args[1]
	}
	if args[2] == productName {
		s.ProductName = args[3]
	}
	if args[2] == jobName {
		s.JobName = args[3]
	}
}

//Backup - method to execute backup
func (s *NFSPlugin) Backup() (err error) {
	lo.G.Info("starting backup for ", s.ProductName, s.JobName)
	var sshConfigs []command.SshConfig
	var writer io.WriteCloser
	if sshConfigs, err = s.getSSHConfig(); err == nil {
		//take first node to execute restore on
		sshConfig := sshConfigs[0]
		if writer, err = s.PivotalCF.NewArchiveWriter(fmt.Sprintf(outputFileName, s.ProductName, s.JobName)); err == nil {
			defer writer.Close()
			s.Dump(sshConfig, writer)
		}
	}
	lo.G.Info("done backup", err)
	return
}

//Restore - method to execute restore
func (s *NFSPlugin) Restore() (err error) {
	lo.G.Info("starting restore for ", s.ProductName, s.JobName)
	var sshConfigs []command.SshConfig
	var reader io.ReadCloser
	if sshConfigs, err = s.getSSHConfig(); err == nil {
		//take first node to execute restore on
		sshConfig := sshConfigs[0]
		if reader, err = s.PivotalCF.NewArchiveReader(fmt.Sprintf(outputFileName, s.ProductName, s.JobName)); err == nil {
			defer reader.Close()
			s.Import(sshConfig, reader)
		}
	}
	lo.G.Info("done restore", err)
	return
}

//GetSSHConfig -
func (s *NFSPlugin) getSSHConfig() (sshConfig []command.SshConfig, err error) {
	var IPs []string
	var vmCredentials cfbackup.VMCredentials

	if IPs, err = s.InstallationSettings.FindIPsByProductAndJob(s.ProductName, s.JobName); err == nil {
		if vmCredentials, err = s.InstallationSettings.FindVMCredentialsByProductAndJob(s.ProductName, s.JobName); err == nil {
			for _, ip := range IPs {
				sshConfig = append(sshConfig, command.SshConfig{
					Username: vmCredentials.UserID,
					Password: vmCredentials.Password,
					Host:     ip,
					Port:     defaultSSHPort,
					SSLKey:   vmCredentials.SSLKey,
				})
			}
		}
	}
	return
}

const (
	pluginName               = "nfs-tile"
	outputFileName           = pluginName + "-%s-%s.dmp"
	defaultSSHPort       int = 22
	nfsRemoteArchivePath     = "/var/vcap/store/shared/archive.backup"
)

//NewNFSPlugin - Contructor helper
func NewNFSPlugin() *NFSPlugin {
	NFSPlugin := &NFSPlugin{
		Meta: cfopsplugin.Meta{
			Name: pluginName,
		},
		Dump:   dumpNFS,
		Import: importNFS,
	}
	return NFSPlugin
}

//NFSPlugin - structure
type NFSPlugin struct {
	PivotalCF            cfopsplugin.PivotalCF
	InstallationSettings cfbackup.InstallationSettings
	Meta                 cfopsplugin.Meta
	ProductName          string
	JobName              string
	Dump                 func(config command.SshConfig, writer io.WriteCloser) (err error)
	Import               func(config command.SshConfig, reader io.ReadCloser) (err error)
}

func dumpNFS(config command.SshConfig, writer io.WriteCloser) (err error) {
	var backup *cfbackup.NFSBackup
	if backup, err = cfbackup.NewNFSBackup(config.Password, config.Host, config.SSLKey, nfsRemoteArchivePath); err == nil {
		lo.G.Debug("Starting nfs dump")
		err = backup.Dump(writer)
		lo.G.Debug("Dump finished", err)
	}
	return
}

func importNFS(config command.SshConfig, reader io.ReadCloser) (err error) {
	var backup *cfbackup.NFSBackup
	if backup, err = cfbackup.NewNFSBackup(config.Password, config.Host, config.SSLKey, nfsRemoteArchivePath); err == nil {
		lo.G.Debug("Starting nfs restore")
		err = backup.Import(reader)
		lo.G.Debug("Restore finished", err)
	}
	return
}
