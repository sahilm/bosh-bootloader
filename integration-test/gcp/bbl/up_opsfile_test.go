package integration_test

import (
	"fmt"

	integration "github.com/cloudfoundry/bosh-bootloader/integration-test"
	"github.com/cloudfoundry/bosh-bootloader/integration-test/actors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("up with opsfile test", func() {
	var (
		bbl       actors.BBL
		gcp       actors.GCP
		terraform actors.Terraform
		boshcli   actors.BOSHCLI
		state     integration.State
	)

	BeforeEach(func() {
		var err error
		configuration, err := integration.LoadConfig()
		Expect(err).NotTo(HaveOccurred())

		state = integration.NewState(configuration.StateFileDir)
		bbl = actors.NewBBL(configuration.StateFileDir, pathToBBL, configuration, "opsfile-env")
		gcp = actors.NewGCP(configuration)
		terraform = actors.NewTerraform(configuration)
		boshcli = actors.NewBOSHCLI()
	})

	It("successfully bbls up and destroys", func() {
		var (
			expectedSSHKey  string
			directorAddress string
			caCertPath      string
			urlToSSLCert    string
		)

		By("calling bbl up with an opsfile", func() {
			bbl.Up(actors.GCPIAAS, []string{"--name", bbl.PredefinedEnvID(), "--ops-file", "some-ops-file"})
		})

		By("checking the ssh key exists", func() {
			expectedSSHKey = fmt.Sprintf("vcap:%s vcap", state.SSHPublicKey())

			actualSSHKeys, err := gcp.SSHKey()
			Expect(err).NotTo(HaveOccurred())
			Expect(actualSSHKeys).To(ContainSubstring(expectedSSHKey))
		})

		By("checking that the bosh director exists", func() {
			directorAddress = bbl.DirectorAddress()
			caCertPath = bbl.SaveDirectorCA()
			exists, err := boshcli.DirectorExists(directorAddress, caCertPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})

		By("checking that the user opsfile was applied", func() {
			manifest, err := boshcli.Manifest(address, caCertPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(manifest).To(ContainSubstring("some-ops"))
		})

		By("calling bbl up", func() {
			bbl.Up(actors.GCPIAAS, []string{})
		})

		By("checking that the user opsfile is still applied", func() {
			manifest, err := boshcli.Manifest(address, caCertPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(manifest).To(ContainSubstring("some-ops"))
		})

		By("calling bbl up with another opsfile", func() {
			bbl.Up(actors.GCPIAAS, []string{"--ops-file", "some-other-ops-file"})
		})

		By("checking that the new opsfile is also applied", func() {
			manifest, err := boshcli.Manifest(address, caCertPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(manifest).To(ContainSubstring("some-ops"))
			Expect(manifest).To(ContainSubstring("some-other-ops"))
		})

		By("calling bbl destroy", func() {
			bbl.Destroy()
		})

		By("checking that the bosh director does not exist", func() {
			exists, _ := boshcli.DirectorExists(directorAddress, caCertPath)
			Expect(exists).To(BeFalse())
		})
	})
})
