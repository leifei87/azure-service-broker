# Service Broker Test App

This is a very naive [Golang](https://golang.org) demo application that demonstrates how to use bindings given by the [Azure Service Broker](https://github.com/azure/meta-azure-service-broker).

## To run the application on Cloud Foundry

1. Log in to your Cloud Foundry using the `cf login` command.

1. From the main project directory, push your app to Cloud Foundry using the `cf push` command. Take note of the route to your app.

1. Create a new service or bind to an existing one, and hit the test-<servicename> endpoint to test your bindings. For example:

## Endpoint list

* `/test-storage` - lists all containers in the given storage account
