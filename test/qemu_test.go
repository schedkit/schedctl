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
	err := runInQemu(t, "../internal/containerd/containerd_test.go")
	if err != nil {
		t.Fatalf("Error running containerd tests in QEMU: %s", err)
	}

	err = runInQemu(t, "../internal/podman/podman_test.go")
	if err != nil {
		t.Fatalf("Error running Podman tests in QEMU: %s", err)
	}
}

func runInQemu(t *testing.T, testPath string) error {
	cmd := exec.Command("go", "test", "-tags", "containers_image_openpgp", "-c", testPath, "-o", "qemu_run_test")
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	if testing.Verbose() {
		log.Print("compile in-qemu test binary")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	err := cmd.Run()
	if err != nil {
		return err
	}
	defer os.Remove("qemu_run_test")

	disk := vmtest.QemuDisk{
		Path:   "../testdata/rootfs.cow",
		Format: "qcow2",
	}

	opts := vmtest.QemuOptions{
		OperatingSystem: vmtest.OS_LINUX,
		Kernel:          "../testdata/bzImage",
		Params:          []string{"-netdev", "user,id=net0,hostfwd=tcp::10022-:22", "-device", "e1000,netdev=net0", "-enable-kvm", "-cpu", "host", "-m", "2048M"},
		Disks: []vmtest.QemuDisk{
			disk,
		},
		Append:  []string{"root=/dev/sda", "rw", "ip=10.0.2.15::10.0.2.2:255.255.255.0::eth0:off"},
		Verbose: testing.Verbose(),
		Timeout: 120 * time.Second,
	}
	// Run QEMU instance
	qemu, err := vmtest.NewQemu(&opts)
	if err != nil {
		return err
	}
	// Shutdown QEMU at the end of the test case
	defer qemu.Shutdown()

	config := &ssh.ClientConfig{
		User:            "root",
		Auth:            []ssh.AuthMethod{ssh.Password("")},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
	}

	// Retry SSH connection until the VM has finished booting
	var conn *ssh.Client
	for i := range 20 {
		conn, err = ssh.Dial("tcp", "127.0.0.1:10022", config)
		if err == nil {
			break
		}
		if i == 19 {
			return fmt.Errorf("ssh connection failed after retries: %w", err)
		}
		time.Sleep(time.Second)
	}
	defer conn.Close()

	// Run a trivial command to ensure SSH subsystem is fully ready
	warmup, err := conn.NewSession()
	if err != nil {
		return fmt.Errorf("creating warmup session: %w", err)
	}
	if _, err := warmup.CombinedOutput("true"); err != nil {
		return fmt.Errorf("warmup command failed: %w", err)
	}

	scpSess, err := conn.NewSession()
	if err != nil {
		return fmt.Errorf("creating scp session: %w", err)
	}

	err = scp.CopyPath("qemu_run_test", "qemu_run_test", scpSess)
	if err != nil {
		return fmt.Errorf("scp failed: %w", err)
	}

	sess, err := conn.NewSession()
	if err != nil {
		return fmt.Errorf("creating test session: %w", err)
	}
	defer sess.Close()

	testCmd := "chmod +x qemu_run_test && ./qemu_run_test"
	if testing.Verbose() {
		testCmd += " -test.v"
	}

	output, err := sess.CombinedOutput(testCmd)
	t.Logf("test output:\n%s", output)
	if err != nil {
		return fmt.Errorf("test failed: %w", err)
	}

	return nil
}
