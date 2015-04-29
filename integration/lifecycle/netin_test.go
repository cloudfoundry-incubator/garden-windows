package lifecycle_test

import (
	"bytes"
	"fmt"
	"os"

	"github.com/cloudfoundry-incubator/garden"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("NetIn", func() {
	var c garden.Container
	var err error

	JustBeforeEach(func() {
		client = startGarden()
		c, err = client.Create(garden.ContainerSpec{})
		Expect(err).ToNot(HaveOccurred())

		// Put show_port.bat in container
		tarFile, err := os.Open("../bin/show_port.tgz")
		Expect(err).ShouldNot(HaveOccurred())
		defer tarFile.Close()
		err = c.StreamIn("bin", tarFile)
		Expect(err).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
		err := client.Destroy(c.Handle())
		Expect(err).ShouldNot(HaveOccurred())
	})

	It("overrides the container's internal port with it's external port", func() {
		By("Creating two NetIn mappings")
		const externalPort1, internalPort1 uint32 = 1000, 1001
		_, _, err := c.NetIn(externalPort1, internalPort1)
		Expect(err).ShouldNot(HaveOccurred())

		const externalPort2, internalPort2 uint32 = 2000, 2001
		_, _, err = c.NetIn(externalPort2, internalPort2)
		Expect(err).ShouldNot(HaveOccurred())

		By("Mapping 1's container port is substituted for it's external port")
		stdout := bytes.NewBuffer(make([]byte, 0, 1024*1024))
		process, err := c.Run(garden.ProcessSpec{
			Path: "bin/show_port.bat",
			Env:  []string{fmt.Sprintf("PORT=%v", internalPort1)},
		}, garden.ProcessIO{Stdout: stdout})
		Expect(err).ShouldNot(HaveOccurred())
		_, err = process.Wait()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(stdout).Should(ContainSubstring(fmt.Sprintf("PORT=%v", externalPort1)))

		By("Mapping 2's container port is substituted for it's external port")
		stdout = bytes.NewBuffer(make([]byte, 0, 1024*1024))
		process, err = c.Run(garden.ProcessSpec{
			Path: "bin/show_port.bat",
			Env:  []string{fmt.Sprintf("PORT=%v", internalPort2)},
		}, garden.ProcessIO{Stdout: stdout})
		Expect(err).ShouldNot(HaveOccurred())
		_, err = process.Wait()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(stdout).Should(ContainSubstring(fmt.Sprintf("PORT=%v", externalPort2)))

		By("Mapping 2's container arg port sets its env port from the mapping")
		stdout = bytes.NewBuffer(make([]byte, 0, 1024*1024))
		process, err = c.Run(garden.ProcessSpec{
			Path: "bin/show_port.bat",
			Args: []string{fmt.Sprintf("-port=%v", internalPort2)},
		}, garden.ProcessIO{Stdout: stdout})
		Expect(err).ShouldNot(HaveOccurred())
		_, err = process.Wait()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(stdout).Should(ContainSubstring(fmt.Sprintf("PORT=%v", externalPort2)))
	})
})