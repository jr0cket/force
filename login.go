package main

import (
	"encoding/json"
	"fmt"
	"net/url"
)

var cmdLogin = &Command{
	Usage: "login",
	Short: "force login [-i=<instance>] [<-u=username> <-p=password>]",
	Long: `
  force login [-i=<instance>] [<-u=username> <-p=password>]

  Examples:
    force login                     
    force login -i=test             
    force login -u=un -p=pw         
    force login -i=test -u=un -p=pw 
    force login -i=na1-blitz01.soma.salesforce.com -u=un -p=pw 
`,
}

func init() {
	cmdLogin.Run = runLogin
}

var (
	instance = cmdLogin.Flag.String("i", "", "non-production server to login to (values are 'pre', 'test', or full instance url")
	userName = cmdLogin.Flag.String("u", "", "Username for Soap Login")
	password = cmdLogin.Flag.String("p", "", "Password for Soap Login")
)

func runLogin(cmd *Command, args []string) {
	var endpoint ForceEndpoint = EndpointProduction

	switch *instance {
	case "test":
		endpoint = EndpointTest
	case "pre":
		endpoint = EndpointPrerelease
	default:
		if *instance != "" {
			//need to determine the form of the endpoint
			uri, err := url.Parse(*instance)
			if err != nil {
				ErrorAndExit("no such endpoint: %s", *instance)
			}
			// Could be short hand?
			if uri.Host == "" {
				uri, err = url.Parse(fmt.Sprintf("https://%s", *instance))
				//fmt.Println(uri)
				if err != nil {
					ErrorAndExit("no such endpoint: %s", *instance)
				}
			}
			CustomEndpoint = uri.Scheme + "://" + uri.Host
			endpoint = EndpointCustom
		}
	}

	if *userName != "" && *password != "" { // Do SOAP login
		_, err := ForceLoginAndSaveSoap(endpoint, *userName, *password)
		if err != nil {
			ErrorAndExit(err.Error())
		}
	} else { // Do OAuth login
		_, err := ForceLoginAndSave(endpoint)
		if err != nil {
			ErrorAndExit(err.Error())
		}
	}
}

func ForceSaveLogin(creds ForceCredentials) (username string, err error) {
	force := NewForce(creds)
	login, err := force.Get(creds.Id)
	if err != nil {
		return
	}
	body, err := json.Marshal(creds)
	if err != nil {
		return
	}
	username = login["username"].(string)

	describe, err := force.Metadata.DescribeMetadata()
	creds.Namespace = describe.NamespacePrefix

	Config.Save("accounts", username, string(body))
	Config.Save("current", "account", username)
	return
}

func ForceLoginAndSaveSoap(endpoint ForceEndpoint, user_name string, password string) (username string, err error) {
	creds, err := ForceSoapLogin(endpoint, user_name, password)
	if err != nil {
		return
	}

	username, err = ForceSaveLogin(creds)
	return
}

func ForceLoginAndSave(endpoint ForceEndpoint) (username string, err error) {
	creds, err := ForceLogin(endpoint)
	if err != nil {
		return
	}
	username, err = ForceSaveLogin(creds)
	return
}
