package main

import (
	"encoding/json"
	"net/http"
	"os"
	"fmt"

	"github.com/pivotal-golang/lager"
	"github.com/gorilla/mux"
	"github.com/Azure/azure-sdk-for-go/storage"
)

var (
	accountName string
	accountKey  string
	blobCli     storage.BlobStorageClient
)

func AttachRoutes(router *mux.Router) {
	router.HandleFunc("/test-storage", listContainers).Methods("GET")
}

func NewAppRouter() http.Handler {
	r := mux.NewRouter()
	AttachRoutes(r)
	return r
}

// gets an authenticated config object
func getConfig(serviceName string) map[string]interface{} {
	vcap := os.Getenv("VCAP_SERVICES")

	var asMap map[string][]map[string]interface{}
	if err := json.Unmarshal([]byte(vcap), &asMap); err != nil {
		println("ERROR setting config from json")
		println(err.Error())
		println(vcap)
	}

	credsInterface := asMap[serviceName][0]["credentials"]

	return credsInterface.(map[string]interface{})
}

func listContainers(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Azure Storage Sample")

	storageCredentials := getConfig("azure-storage")
	accountName := storageCredentials["storage_account_name"].(string)
	accountKey := storageCredentials["primary_access_key"].(string)
	client, err := storage.NewBasicClient(accountName, accountKey)
	onErrorFail(err, "Create client failed")

	blobCli = client.GetBlobService()

	fmt.Println("Create container with private access type...")
	containerName := "test061901"
	cnt := blobCli.GetContainerReference(containerName)
	options := storage.CreateContainerOptions{
		Access: storage.ContainerAccessTypePrivate,
	}
	_, err = cnt.CreateIfNotExists(&options)

	res, err := blobCli.ListContainers(storage.ListContainersParameters{})
	containers := res.Containers
	println("encoding response", err, lager.Data{"response": res})

	respond(w, http.StatusOK, containers)
}

func respond(w http.ResponseWriter, status int, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	encoder := json.NewEncoder(w)
	err := encoder.Encode(response)
	if err != nil {
		println("encoding response", err, lager.Data{"status": status, "response": response})
	}
}

// onErrorFail prints a failure message and exits the program if err is not nil.
func onErrorFail(err error, message string) {
	if err != nil {
		fmt.Printf("%s: %s\n", message, err)
		os.Exit(1)
	}
}
