package os

func init() {
	RegisterRuntime("ubuntu", &RegisteredRuntime{
		Detect: UbuntuDetection,
	})
}

func UbuntuDetection() (*Runtime, error) {
	return nil, nil
}

type Ubuntu struct{}

func (r *Ubuntu) Service(name string, state ServiceState) error {
	return nil
}

func (r *Ubuntu) Package(name string, state PackageState) error {
	return nil
}
