# Godo

Godo is a Go client library for accessing the DigitalOcean V2 API.

You can view the client API docs here: [http://godoc.org/github.com/digitalocean/godo](http://godoc.org/github.com/digitalocean/godo)

You can view Digital Ocean API docs here: [https://developers.digitalocean.com/v2/](https://developers.digitalocean.com/v2/)


## Usage

```go
import "github.com/digitalocean/godo"
```

Create a new DigitalOcean client, then use the exposed services to
access different parts of the DigitalOcean API.

### Authentication

Currently, Personal Access Token (PAT) is the only method of
authenticating with the API. You can manage your tokens
at the Digital Ocean Control Panel [Applications Page](https://cloud.digitalocean.com/settings/applications).

You can then use your token to create a new client:

```go
import "code.google.com/p/goauth2/oauth"

pat := "mytoken"
t := &oauth.Transport{
	Token: &oauth.Token{AccessToken: pat},
}

client := godo.NewClient(t.Client())
```

## Examples


To create a new Droplet:

```go
dropletName := "super-cool-droplet"

createRequest := &godo.DropletCreateRequest{
    Name:   dropletName,
    Region: "nyc3",
    Size:   "512mb",
    Image:  "ubuntu-14-04-x64",
}

newDroplet, _, err := client.Droplets.Create(createRequest)

if err != nil {
    fmt.Printf("Something bad happened: %s\n\n", err)
    return err
}
```

### Pagination

If a list of items is paginated by the API, you must request pages individually. For example, to fetch all Droplets:

```go
func DropletList(client *godo.Client) ([]godo.Droplet, error) {
    // create a list to hold our droplets
    list := []godo.Droplet{}

    // create options. initially, these will be blank
    opt := &godo.ListOptions{}
    for {
        droplets, resp, err := client.Droplets.List(opt)
        if err != nil {
            return err
        }
        
        // append the current page's droplets to our list
        for _, d := range droplets {
            list = append(list, d)
        }

       // if we are at the last page, break out the for loop
       if resp.Links.IsLastPage() {
           break
       }

       page, err := resp.Links.CurrentPage()
       if err != nil {
           return err
        }

       // set the page we want for the next request
       opt.Page = page + 1
    }

    return nil
}

```
