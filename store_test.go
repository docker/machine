package machine

import (
	"testing"

	_ "github.com/docker/machine/drivers/none"
)

func TestStoreList(t *testing.T) {
	if err := cleanup(); err != nil {
		t.Fatal(err)
	}

	st := NewStore(TestStoreDir, "", "")

	m, err := getTestMachine()
	if err != nil {
		t.Fatal(err)
	}

	if err := st.Save(m); err != nil {
		t.Fatal(err)
	}

	machines, errs := st.List()
	if len(errs) != 0 {
		t.Fatal(errs)
	}

	if len(machines) != 1 {
		t.Fatalf("list returned %d items", len(machines))
	}

	if machines[0].Name != "test" {
		t.Fatalf("machines[0] name is incorrect, got: %s", machines[0].Name)
	}

	if err := cleanup(); err != nil {
		t.Fatal(err)
	}
}

func TestStoreExists(t *testing.T) {
	if err := cleanup(); err != nil {
		t.Fatal(err)
	}

	st := NewStore(TestStoreDir, "", "")

	m, err := getTestMachine()
	if err != nil {
		t.Fatal(err)
	}

	if err := st.Save(m); err != nil {
		t.Fatal(err)
	}

	exists, err := st.Exists("test")
	if err != nil {
		t.Fatal(err)
	}

	if !exists {
		t.Fatal("exists returned false when it should have been true")
	}

	if err := cleanup(); err != nil {
		t.Fatal(err)
	}
}

func TestStoreLoad(t *testing.T) {
	if err := cleanup(); err != nil {
		t.Fatal(err)
	}

	st := NewStore(TestStoreDir, "", "")

	m, err := getTestMachine()
	if err != nil {
		t.Fatal(err)
	}

	if err := st.Save(m); err != nil {
		t.Fatal(err)
	}

	exists, err := st.Exists("test")
	if err != nil {
		t.Fatal(err)
	}

	if !exists {
		t.Fatal("exists returned false when it should have been true")
	}

	expectedUrl := "unix:///var/run/docker.sock"

	machine, err := st.Get("test")
	if err != nil {
		t.Fatal(err)
	}

	if machine.Name != "test" {
		t.Fatal("Host name is incorrect")
	}

	actualUrl, err := machine.GetURL()
	if err != nil {
		t.Fatal(err)
	}

	if actualUrl != expectedUrl {
		t.Fatalf("expected url %q, received %q", expectedUrl, actualUrl)
	}

	if err := cleanup(); err != nil {
		t.Fatal(err)
	}
}

func TestStoreGetSetActive(t *testing.T) {
	if err := cleanup(); err != nil {
		t.Fatal(err)
	}

	st := NewStore(TestStoreDir, "", "")

	// No machine set
	machine, err := st.GetActive()
	if err != nil {
		t.Fatal(err)
	}

	if machine != nil {
		t.Fatalf("GetActive: active machine should not exist")
	}

	// Set normal machine
	originalMachine, err := getTestMachine()
	if err != nil {
		t.Fatal(err)
	}

	if err := st.Save(originalMachine); err != nil {
		t.Fatal(err)
	}

	if err := st.SetActive(originalMachine); err != nil {
		t.Fatal(err)
	}

	machine, err = st.GetActive()
	if err != nil {
		t.Fatal(err)
	}

	if machine.Name != "test" {
		t.Fatalf("active machine is not 'test', got %s", machine.Name)
	}

	isActive, err := st.IsActive(machine)
	if err != nil {
		t.Fatal(err)
	}

	if isActive != true {
		t.Fatal("IsActive: active machine is not test")
	}

	// remove active machine
	if err := st.RemoveActive(); err != nil {
		t.Fatal(err)
	}

	machine, err = st.GetActive()
	if err != nil {
		t.Fatal(err)
	}

	if machine != nil {
		t.Fatalf("active machine %s is not nil", machine.Name)
	}

	if err := cleanup(); err != nil {
		t.Fatal(err)
	}
}
