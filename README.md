# Building and Testing a REST API in GoLang using Gorilla Mux and MySQL

## Prerequisites

1. Create a [Cloud SQL for MySQL Second Generation](https://cloud.google.com/go/getting-started/using-cloud-sql#create_a_cloud_sql_instance) instance in GCP. Write down the associated username and password (initially `root` but feel free to create a dedicated user).
1. Create a database, for example `rest_api_example`.
1. Write down the connection name, in the format `[PROJECT_NAME]:[REGION_NAME]:[INSTANCE_NAME]` (for example `devops-terraform-deployer:us-central1:rest-api-example`).
1. For local development, download and install the [Cloud SQL Proxy](https://cloud.google.com/go/getting-started/using-cloud-sql#install_the_cloud_sql_proxy). For example, for MacOS 64-bit:
    ```
    $ curl -o cloud_sql_proxy https://dl.google.com/cloudsql/cloud_sql_proxy.darwin.amd64
    $ chmod +x cloud_sql_proxy
    ```

## Local Development

1. Start the Cloud SQL Proxy using the connection name from the previous step.
    ```
    $ ./cloud_sql_proxy -instances="devops-terraform-deployer:us-central1:store"=tcp:3306
    2019/05/07 20:25:15 current FDs rlimit set to 12800, wanted limit is 8500. Nothing to do here.
    2019/05/07 20:25:15 Listening on 127.0.0.1:3306 for devops-terraform-deployer:us-central1:store
    2019/05/07 20:25:15 Ready for new connections
    ```
1. Create environment variables for `DB_USERNAME`, `DB_PASSWORD`, `DB_INSTANCE_NAME`, `DB_NAME` and `PORT` (used in `main.go` and `main_test.go`) with:
    ```
    export DB_USERNAME=...
    export DB_PASSWORD=...
    export DB_INSTANCE_NAME=devops-terraform-deployer:us-central1:store
    export DB_NAME=rest_api_example
    export PORT=8080
    ```
1. Run the tests with:
    ```
    $ go test -v
    === RUN   TestEmptyTable
    --- PASS: TestEmptyTable (0.34s)
    === RUN   TestGetNonExistentUser
    --- PASS: TestGetNonExistentUser (0.34s)
    === RUN   TestCreateUser
    --- PASS: TestCreateUser (0.46s)
    === RUN   TestGetUser
    --- PASS: TestGetUser (0.47s)
    === RUN   TestUpdateUser
    --- PASS: TestUpdateUser (0.59s)
    === RUN   TestDeleteUser
    --- PASS: TestDeleteUser (0.71s)
    PASS
    ok      github.com/lvaylet/go-chi-mysql 3.763s
    ```
1. Build and run the app with:
    ```
    $ go build
    $ ./go-chi-mysql
    ```
1. Open another terminal and query the REST API with:
    ```
    $ curl http://localhost:8080/users
    []
    $ curl -X POST -d '{"name": "John", "age": 30}' http://localhost:8080/user
    {"id":1,"name":"John","age":30}
    $ curl http://localhost:8080/users
    [{"id":1,"name":"John","age":30}]
    $ curl http://localhost:8080/user/1
    {"id":1,"name":"John","age":30}
    $ curl http://localhost:8080/user/1
    {"id":1,"name":"John","age":30}
    $ curl -X PUT -d '{"name": "John", "age": 32}' http://localhost:8080/user/1
    {"id":1,"name":"John","age":32}
    $ curl -X DELETE http://localhost:8080/user/1
    {"result":"success"}
    $ curl http://localhost:8080/user/1
    {"error":"User not found"}
    $ curl http://localhost:8080/users
    []
    ```

## Deployment to Google Cloud Run

1. Build and deploy the Docker image with:
    ```
    export PROJECT=<YOUR_GCP_PROJECT_ID>
    export REGION=us-central1
    export CLOUD_RUN_SVC=go-chi-mysql
    gcloud config set project ${PROJECT}
    gcloud config set run/region ${REGION}
    gcloud builds submit --tag gcr.io/${PROJECT}/${CLOUD_RUN_SVC}
    gcloud beta run deploy ${CLOUD_RUN_SVC} \
        --image gcr.io/${PROJECT}/${CLOUD_RUN_SVC} \
        --allow-unauthenticated \
        --update-env-vars DB_USERNAME=$DB_USERNAME,DB_PASSWORD=$DB_PASSWORD,DB_INSTANCE_NAME=$DB_INSTANCE_NAME,DB_NAME=$DB_NAME
    ````
1. Retrieve the exposed endpoint with:
    ```
    CLOUD_RUN_ENDPOINT=$(gcloud beta run services describe ${CLOUD_RUN_SVC} --format "value(status.address.hostname)")
    ```
1. Query the endpoint, for example with:
    ```
    $ curl ${CLOUD_RUN_ENDPOINT}/users
    []
    ```

## References

- https://medium.com/@kelvin_sp/building-and-testing-a-rest-api-in-golang-using-gorilla-mux-and-mysql-1f0518818ff6
- https://semaphoreci.com/community/tutorials/building-and-testing-a-rest-api-in-go-with-gorilla-mux-and-postgresql
- https://cloud.google.com/go/getting-started/using-cloud-sql
- https://github.com/lvaylet/go-chi-rest-api
- https://thenewstack.io/tutorial-deploying-a-web-application-on-google-cloud-run/