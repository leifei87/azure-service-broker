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
	"github.com/Azure/azure-sdk-for-go/arm/resources/resources"	
	"github.com/Azure/azure-sdk-for-go/arm/redis"
	"github.com/Azure/azure-sdk-for-go/arm/documentdb"
	"github.com/Azure/azure-sdk-for-go/arm/servicebus"
	"github.com/Azure/azure-sdk-for-go/arm/sql"
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
	router.HandleFunc("/test-documentdb", getDocumentDB).Methods("GET")
	router.HandleFunc("/test-servicebus", getServiceBus).Methods("GET")
	router.HandleFunc("/test-sqldb", getSQLDB).Methods("GET")
	router.HandleFunc("/test-mysqldb", getmySQLDB).Methods("GET")
	router.HandleFunc("/test-postgresqldb", getpostgresqlDB).Methods("GET")
	router.HandleFunc("/test-cosmosdb", getCosmosDB).Methods("GET")
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
	containerName := getEnvVarOrExit("STORAGE_CONTAINER_NAME")
	client, err := storage.NewBasicClient(accountName, accountKey)
	onErrorFail(err, "Create client failed")

	blobCli = client.GetBlobService()

	fmt.Println("Create container with private access type...")
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

func getDocumentDB(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Azure DocumentDB Sample")
	subscriptionID := getEnvVarOrExit("AZURE_SUBSCRIPTION_ID")
	docDBResourceGroup := getEnvVarOrExit("DOCUMENTDB_RESOURGE_GROUP")
	docDBAccountName := getEnvVarOrExit("DOCUMENTDB_ACCOUNT_NAME")

	spt := getServicePrincipalToken()
	client := documentdb.NewDatabaseAccountsClient(subscriptionID)
	client.Authorizer = autorest.NewBearerAuthorizer(spt)

	op,err := client.Get(docDBResourceGroup,docDBAccountName)
	onErrorFail(err, "Get DocumentDB account failed")
	account := *(op.Name)
	println("encoding response", err, lager.Data{"response": op})
	respond(w, http.StatusOK, account)
}

func getSQLDB(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Azure SQL DB Sample")
	subscriptionID := getEnvVarOrExit("AZURE_SUBSCRIPTION_ID")
	sqlDBResourceGroup := getEnvVarOrExit("SQL_RESOURCE_GROUP")
	sqlServerName := getEnvVarOrExit("SQL_SERVER_NAME")
	sqlDBName := getEnvVarOrExit("SQL_DATABASE_NAME")

	spt := getServicePrincipalToken()
	client := sql.NewDatabasesClient(subscriptionID)
	client.Authorizer = autorest.NewBearerAuthorizer(spt)

	op,err := client.Get(sqlDBResourceGroup,sqlServerName,sqlDBName,"serviceTierAdvisors, transparentDataEncryption")
	onErrorFail(err, "Get SQL DB failed")
	sqlDB := *(op.Name)
	println("encoding response", err, lager.Data{"response": op})
	respond(w, http.StatusOK, sqlDB)
}

func getmySQLDB(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Azure SQL DB Sample")
	subscriptionID := getEnvVarOrExit("AZURE_SUBSCRIPTION_ID")
	mysqlDBResourceGroup := getEnvVarOrExit("MYSQL_RESOURCE_GROUP")

	spt := getServicePrincipalToken()
	client := resources.NewGroupsClient(subscriptionID)
	client.Authorizer = autorest.NewBearerAuthorizer(spt)

	var top int32
	top = 1
	op,err := client.ListResources(mysqlDBResourceGroup,"","",&top)
	onErrorFail(err, "Get mySQL failed")
	resourceList := *(op.Value)
	mySQLServerName := *(resourceList[0].Name)
	println("encoding response", err, lager.Data{"response": op})
	respond(w, http.StatusOK, mySQLServerName)
}

func getpostgresqlDB(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Azure postgresql DB Sample")
	subscriptionID := getEnvVarOrExit("AZURE_SUBSCRIPTION_ID")
	postgresqlDBResourceGroup := getEnvVarOrExit("POSTGRESQL_RESOURCE_GROUP")

	spt := getServicePrincipalToken()
	client := resources.NewGroupsClient(subscriptionID)
	client.Authorizer = autorest.NewBearerAuthorizer(spt)

	var top int32
	top = 1
	op,err := client.ListResources(postgresqlDBResourceGroup,"","",&top)
	onErrorFail(err, "Get postgresql failed")
	resourceList := *(op.Value)
	postgreSQLServerName := *(resourceList[0].Name)
	println("encoding response", err, lager.Data{"response": op})
	respond(w, http.StatusOK, postgreSQLServerName)
}

func getCosmosDB(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Azure CosmosDB Sample")
	subscriptionID := getEnvVarOrExit("AZURE_SUBSCRIPTION_ID")
	cosmosDBResourceGroup := getEnvVarOrExit("COSMOSDB_RESOURCE_GROUP")
	cosmosDBAccountName := getEnvVarOrExit("COSMOSDB_ACCOUNT_NAME")

	spt := getServicePrincipalToken()
	client := documentdb.NewDatabaseAccountsClient(subscriptionID)
	client.Authorizer = autorest.NewBearerAuthorizer(spt)

	op,err := client.Get(cosmosDBResourceGroup,cosmosDBAccountName)
	onErrorFail(err, "Get cosmosDB account failed")
	account := *(op.Name)
	println("encoding response", err, lager.Data{"response": op})
	respond(w, http.StatusOK, account)
}

func getServiceBus(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Azure Service Bus Sample")
	subscriptionID := getEnvVarOrExit("AZURE_SUBSCRIPTION_ID")
	sbResourceGroup := getEnvVarOrExit("SERVICEBUS_RESOURGE_GROUP")
	sbName := getEnvVarOrExit("SERVICEBUS_NAME")
	queueName := getEnvVarOrExit("SERVICEBUS_QUEUE_NAME")
	location := "eastus"
	parameter := servicebus.QueueCreateOrUpdateParameters {
		Location: &location,
	}

	spt := getServicePrincipalToken()
	queueclient := servicebus.NewQueuesClient(subscriptionID)
	queueclient.Authorizer = autorest.NewBearerAuthorizer(spt)

	op,err := queueclient.CreateOrUpdate(sbResourceGroup,sbName,queueName,parameter)
	onErrorFail(err, "Create Queue failed")
	queue := *(op.Name)
	println("encoding response", err, lager.Data{"response": op})
	respond(w, http.StatusOK, queue)
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