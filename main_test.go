package main_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotalservices/cfbackup"
	"github.com/pivotalservices/cfbackup/tileregistry"
	. "github.com/pivotalservices/cfops-nfs-plugin"
	"github.com/pivotalservices/cfops/plugin/cfopsplugin"
	"github.com/pivotalservices/gtils/command"
)

var _ = Describe("Given NFSPlugin", func() {
	var nfsPlugin *NFSPlugin
	Describe("given a Meta() method", func() {
		Context("called on a plugin with valid meta data", func() {
			var meta cfopsplugin.Meta
			BeforeEach(func() {
				nfsPlugin = NewNFSPlugin()
				meta = nfsPlugin.GetMeta()
			})

			It("then it should return a meta data object with all required fields", func() {
				Ω(meta.Name).ShouldNot(BeEmpty())
			})
		})
	})
	testInstallationSettings("./fixtures/installation-settings-nfs.json")
})

func testInstallationSettings(installationSettingsPath string) {
	const pluginArgs = "--productName cf --jobName nfs_server"
	var nfsPlugin *NFSPlugin
	Describe(fmt.Sprintf("given a installationSettingsFile %s", installationSettingsPath), func() {
		Describe("given a Backup() method", func() {
			Context("when called on a properly setup NFSPlugin object", func() {
				var err error
				var controlTmpDir string
				var counter int
				BeforeEach(func() {
					controlTmpDir, _ = ioutil.TempDir("", "unit-test")
					nfsPlugin = &NFSPlugin{
						Meta: cfopsplugin.Meta{
							Name: "nfs-tile",
						},
						Dump: func(config command.SshConfig, writer io.WriteCloser) (err error) {
							counter++
							return
						},
					}
					configParser := cfbackup.NewConfigurationParser(installationSettingsPath)
					pivotalCF := cfopsplugin.NewPivotalCF(configParser.InstallationSettings, tileregistry.TileSpec{
						ArchiveDirectory: controlTmpDir,
						PluginArgs:       pluginArgs,
					})
					nfsPlugin.Setup(pivotalCF)
					err = nfsPlugin.Backup()
				})

				AfterEach(func() {
					os.RemoveAll(controlTmpDir)
				})

				It("then it should have created right number of archive files", func() {
					Ω(err).ShouldNot(HaveOccurred())
					Ω(counter).Should(Equal(1))
				})
			})
		})
		Describe("given a Restore() method", func() {
			Context("when called on a properly setup NFSPlugin object", func() {
				var err error
				var controlTmpDir string
				var counter int
				BeforeEach(func() {
					controlTmpDir, _ = ioutil.TempDir("", "unit-test")
					os.Create(controlTmpDir + "/nfs-tile-cf-nfs_server.dmp")
					nfsPlugin = &NFSPlugin{
						Meta: cfopsplugin.Meta{
							Name: "nfs-tile",
						},
						Import: func(config command.SshConfig, reader io.ReadCloser) (err error) {
							counter++
							return
						},
					}
					configParser := cfbackup.NewConfigurationParser(installationSettingsPath)
					pivotalCF := cfopsplugin.NewPivotalCF(configParser.InstallationSettings, tileregistry.TileSpec{
						ArchiveDirectory: controlTmpDir,
						PluginArgs:       pluginArgs,
					})
					nfsPlugin.Setup(pivotalCF)
					err = nfsPlugin.Restore()
				})

				AfterEach(func() {
					os.RemoveAll(controlTmpDir)
				})

				It("then it should have ran 1 import", func() {
					Ω(err).ShouldNot(HaveOccurred())
					Ω(counter).Should(Equal(1))
				})
			})
		})

		Describe("given a Setup() method", func() {
			Context("when called with a PivotalCF containing a NFS tile", func() {
				var pivotalCF cfopsplugin.PivotalCF
				var configParser *cfbackup.ConfigurationParser
				BeforeEach(func() {
					configParser = cfbackup.NewConfigurationParser(installationSettingsPath)
				})
				It("then it should not panic", func() {
					pivotalCF = cfopsplugin.NewPivotalCF(configParser.InstallationSettings, tileregistry.TileSpec{
						PluginArgs: pluginArgs,
					})
					Ω(func() {
						nfsPlugin.Setup(pivotalCF)
					}).ShouldNot(Panic())
					Ω(nfsPlugin.PivotalCF).ShouldNot(BeNil())
					Ω(nfsPlugin.ProductName).ShouldNot(BeEmpty())
					Ω(nfsPlugin.JobName).ShouldNot(BeEmpty())
				})
				It("then it should panic", func() {
					pivotalCF = cfopsplugin.NewPivotalCF(configParser.InstallationSettings, tileregistry.TileSpec{
						PluginArgs: "",
					})
					Ω(func() {
						nfsPlugin.Setup(pivotalCF)
					}).Should(Panic())
				})

			})
		})
	})
}
