package main

import (
	"encoding/json"
	"net/http"
	"os"
	"fmt"
	"strconv"

	"github.com/pivotal-golang/lager"
	"github.com/gorilla/mux"
	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/Azure/azure-sdk-for-go/arm/redis"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest"
)

var (
	accountName string
	accountKey  string
	blobCli     storage.BlobStorageClient
)

func AttachRoutes(router *mux.Router) {
	router.HandleFunc("/test-storage", listContainers).Methods("GET")
	router.HandleFunc("/test-redis", getRedisCache).Methods("GET")
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

func getRedisCache(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Azure Redis Cache Sample")
	subscriptionID := getEnvVarOrExit("AZURE_SUBSCRIPTION_ID")
	redisResourceGroup := getEnvVarOrExit("REDIS_RESOURGE_GROUP")
	redisName := getEnvVarOrExit("REDIS_NAME")

	spt := getServicePrincipalToken()
	client := redis.NewGroupClient(subscriptionID)
	client.Authorizer = autorest.NewBearerAuthorizer(spt)

	enableport := true
	updateProperties := redis.UpdateProperties{
		EnableNonSslPort: &enableport,
	}
	updateParameters := redis.UpdateParameters{
		&updateProperties,
	}
	op,err := client.Update(redisResourceGroup,redisName,updateParameters)
	onErrorFail(err, "Get Redis failed")
	redis := *(op.Name)
	sslEnabled := *((*(op.ResourceProperties)).EnableNonSslPort)
	result := redis + ":" + strconv.FormatBool(sslEnabled)
	println("encoding response", err, lager.Data{"response": op})

	respond(w, http.StatusOK, result)
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

func getEnvVarOrExit(varName string) string {
		value := os.Getenv(varName)
		if value == "" {
			fmt.Printf("Missing environment variable %s\n", varName)
			os.Exit(1)
		}
		return value
}

func getServicePrincipalToken() *adal.ServicePrincipalToken {
	tenantID := getEnvVarOrExit("AZURE_TENANT_ID")
	oauthConfig, err := adal.NewOAuthConfig(azure.PublicCloud.ActiveDirectoryEndpoint, tenantID)
	onErrorFail(err, "OAuthConfigForTenant failed")
	clientID := getEnvVarOrExit("AZURE_CLIENT_ID")
	clientSecret := getEnvVarOrExit("AZURE_CLIENT_SECRET")
	spt, err := adal.NewServicePrincipalToken(*oauthConfig, clientID, clientSecret, azure.PublicCloud.ResourceManagerEndpoint)
	onErrorFail(err, "NewServicePrincipalToken failed")
	return spt
}