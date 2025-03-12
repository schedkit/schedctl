package integration_test

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/anatol/vmtest"
	"github.com/tmc/scp"
	"golang.org/x/crypto/ssh"
)

func TestIntegrationInQemu(t *testing.T) {
	cmd := exec.Command("go", "test", "-c", "../internal/containerd/containerd_test.go", "-o", "qemu_run_test")
	if testing.Verbose() {
		log.Print("compile in-qemu test binary")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	err := cmd.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("qemu_run_test")

	disk := vmtest.QemuDisk{
		Path:   "../testdata/rootfs.cow",
		Format: "qcow2",
	}

	opts := vmtest.QemuOptions{
		OperatingSystem: vmtest.OS_LINUX,
		Kernel:          "../testdata/bzImage",
		Params:          []string{"-net", "user,hostfwd=tcp::10022-:22", "-net", "nic", "-enable-kvm", "-cpu", "host", "-m", "2048M"},
		Disks: []vmtest.QemuDisk{
			disk,
		},
		Append:  []string{"root=/dev/sda", "rw"},
		Verbose: testing.Verbose(),
		Timeout: 50 * time.Second,
	}
	// Run QEMU instance
	qemu, err := vmtest.NewQemu(&opts)
	if err != nil {
		t.Fatal(err)
	}
	// Shutdown QEMU at the end of the test case
	defer qemu.Shutdown()

	config := &ssh.ClientConfig{
		User:            "root",
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
	}

	conn, err := ssh.Dial("tcp", "localhost:10022", config)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	sess, err := conn.NewSession()
	if err != nil {
		t.Fatal(err)
	}
	defer sess.Close()

	scpSess, err := conn.NewSession()
	if err != nil {
		t.Fatal(err)
	}

	err = scp.CopyPath("qemu_run_test", "qemu_run_test", scpSess)
	if err != nil {
		t.Error(err)
	}

	testCmd := "./qemu_run_test"
	if testing.Verbose() {
		testCmd += " -test.v"
	}

	output, err := sess.CombinedOutput(testCmd)
	if testing.Verbose() {
		fmt.Print(string(output)) //nolint:forbidigo
	}
	if err != nil {
		t.Fatal(err)
	}
}
