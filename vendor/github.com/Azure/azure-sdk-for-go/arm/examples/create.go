package examples

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Azure/azure-sdk-for-go/arm/examples/helpers"
	"github.com/Azure/azure-sdk-for-go/arm/storage"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"
)

func withWatcher() autorest.SendDecorator {
	return func(s autorest.Sender) autorest.Sender {
		return autorest.SenderFunc(func(r *http.Request) (*http.Response, error) {
			fmt.Printf("Sending %s %s\n", r.Method, r.URL)
			resp, err := s.Do(r)
			fmt.Printf("...received status %s\n", resp.Status)
			if autorest.ResponseRequiresPolling(resp) {
				fmt.Printf("...will poll after %d seconds\n",
					int(autorest.GetPollingDelay(resp, time.Duration(0))/time.Second))
				fmt.Printf("...will poll at %s\n", autorest.GetPollingLocation(resp))
			}
			fmt.Println("")
			return resp, err
		})
	}
}

func createAccount(resourceGroup, name string) {
	c, err := helpers.LoadCredentials()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	ac := storage.NewAccountsClient(c["subscriptionID"])

	spt, err := helpers.NewServicePrincipalTokenFromCredentials(c, azure.AzureResourceManagerScope)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	ac.Authorizer = spt

	cna, err := ac.CheckNameAvailability(
		storage.AccountCheckNameAvailabilityParameters{
			Name: to.StringPtr(name),
			Type: to.StringPtr("Microsoft.Storage/storageAccounts")})
	if err != nil {
		log.Fatalf("Error: %v", err)
		return
	}
	if !to.Bool(cna.NameAvailable) {
		fmt.Printf("%s is unavailable -- try again\n", name)
		return
	}
	fmt.Printf("%s is available\n\n", name)

	ac.Sender = autorest.CreateSender(withWatcher())
	ac.PollingMode = autorest.PollUntilAttempts
	ac.PollingAttempts = 5

	cp := storage.AccountCreateParameters{}
	cp.Location = to.StringPtr("westus")
	cp.Properties = &storage.AccountPropertiesCreateParameters{AccountType: storage.StandardLRS}

	sa, err := ac.Create(resourceGroup, name, cp)
	if err != nil {
		if sa.Response.StatusCode != http.StatusAccepted {
			fmt.Printf("Creation of %s.%s failed with err -- %v\n", resourceGroup, name, err)
			return
		}
		fmt.Printf("Create initiated for %s.%s -- poll %s to check status\n",
			resourceGroup,
			name,
			sa.GetPollingLocation())
		return
	}

	fmt.Printf("Successfully created %s.%s\n\n", resourceGroup, name)

	ac.Sender = nil
	r, err := ac.Delete(resourceGroup, name)
	if err != nil {
		fmt.Printf("Delete of %s.%s failed with status %s\n...%v\n", resourceGroup, name, r.Status, err)
		return
	}
	fmt.Printf("Deletion of %s.%s succeeded -- %s\n", resourceGroup, name, r.Status)
}
