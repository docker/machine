package clcgo

type request struct {
	URL        string
	Parameters interface{}
}

// The SavableEntity interface is implemented on any resources that can be
// saved via SaveEntity.
type SavableEntity interface {
	RequestForSave(string) (request, error)
}

// CreationStatusProvidingEntity will be implemented by some SavableEntity
// resources when information about the created resource is not immediately
// available.  For instance, a Server or PublicIPAddress must first be
// provisioned, so a Status object is returned so that you can query it until
// the work has been successfully completed.
//
// All StatusProvidingEntites must be SavableEntities, but not every
// SavableEntity is a CreationStatusProvidingEntity.
type CreationStatusProvidingEntity interface {
	StatusFromCreateResponse([]byte) (Status, error)
}
