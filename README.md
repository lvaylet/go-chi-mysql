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
1. Configure `gcloud` for your project, region and Cloud SQL instance with:
    ```
    export PROJECT=<YOUR_GCP_PROJECT_ID>
    # Cloud Run only supports running in `us-central1` at the time of writing
    export REGION=us-central1
    # Cloud SQL
    export CLOUD_SQL_INSTANCE_ID=...
    export CLOUD_SQL_INSTANCE_CONNECTION_NAME=${PROJECT}:${REGION}:${CLOUD_SQL_INSTANCE_ID}
    export CLOUD_SQL_DB_NAME=...
    export CLOUD_SQL_DB_USERNAME=...
    export CLOUD_SQL_DB_PASSWORD=...
    # Cloud Run
    export CLOUD_RUN_SVC=go-chi-mysql

    gcloud config set project ${PROJECT}
    gcloud config set run/region ${REGION}
    ```

## Local Development

1. Start the Cloud SQL Proxy using the connection name from the previous step.
    ```
    $ ./cloud_sql_proxy -instances="${CLOUD_SQL_INSTANCE_CONNECTION_NAME}"=tcp:3306
    2019/05/07 20:25:15 current FDs rlimit set to 12800, wanted limit is 8500. Nothing to do here.
    2019/05/07 20:25:15 Listening on 127.0.0.1:3306 for <CLOUD_SQL_INSTANCE_CONNECTION_NAME>
    2019/05/07 20:25:15 Ready for new connections
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
    go build
    ./go-chi-mysql
    ```
1. Open another terminal and query the REST API with:
    ```
    $ API_ENDPOINT=https://localhost:8080
    $ curl ${API_ENDPOINT}/users
    []
    $ curl -X POST -d '{"name": "John", "age": 30}' ${API_ENDPOINT}/user
    {"id":1,"name":"John","age":30}
    $ curl ${API_ENDPOINT}/users
    [{"id":1,"name":"John","age":30}]
    $ curl ${API_ENDPOINT}/user/1
    {"id":1,"name":"John","age":30}
    $ curl -X PUT -d '{"name": "John", "age": 32}' ${API_ENDPOINT}/user/1
    {"id":1,"name":"John","age":32}
    $ curl -X DELETE ${API_ENDPOINT}/user/1
    {"result":"success"}
    $ curl ${API_ENDPOINT}/user/1
    {"error":"User not found"}
    $ curl ${API_ENDPOINT}/users
    []
    ```

## Deployment to Google Cloud Run

1. At the time of writing, Cloud Run does not officially support connecting to Cloud SQL instances. The [available workaround](https://medium.com/@petomalina/connecting-to-cloud-sql-from-cloud-run-dcff2e20152a) involves embedding the Cloud SQL Proxy inside the Docker image and granting the Cloud Run Service Account the Cloud SQL Client role. Make sure the Service Account ending with `@serverless-robot-prod.iam.gserviceaccount.com` has the right permissions with:
    ```
    gcloud projects add-iam-policy-binding ${PROJECT} \
        --member serviceAccount:service-<YOUR_SA_ID>@serverless-robot-prod.iam.gserviceaccount.com \
        --role roles/cloudsql.client
    ```
1. Build and push the Docker image with:
    ```
    gcloud builds submit --tag gcr.io/${PROJECT}/${CLOUD_RUN_SVC}
    ```
    Alternatively, you can build locally and deploy with:
    ```
    docker build --tag gcr.io/${PROJECT}/${CLOUD_RUN_SVC} .
    docker push gcr.io/${PROJECT}/${CLOUD_RUN_SVC}
    ```
1. Deploy to Cloud Run with:
    ```
    gcloud beta run deploy ${CLOUD_RUN_SVC} \
        --image gcr.io/${PROJECT}/${CLOUD_RUN_SVC} \
        --allow-unauthenticated \
        --set-env-vars \
            CLOUD_SQL_DB_USERNAME=${CLOUD_SQL_DB_USERNAME},CLOUD_SQL_DB_PASSWORD=${CLOUD_SQL_DB_PASSWORD},CLOUD_SQL_INSTANCE_CONNECTION_NAME=${CLOUD_SQL_INSTANCE_CONNECTION_NAME},CLOUD_SQL_DB_NAME=${CLOUD_SQL_DB_NAME}
    ````
1. Retrieve the exposed endpoint with:
    ```
    API_ENDPOINT=$(gcloud beta run services describe ${CLOUD_RUN_SVC} --format "value(status.address.hostname)")
    ```
1. Query the endpoint, for example with:
    ```
    $ curl ${API_ENDPOINT}/users
    $ curl -X POST -d '{"name": "John", "age": 30}' ${API_ENDPOINT}/user
    $ curl ${API_ENDPOINT}/users
    $ curl ${API_ENDPOINT}/user/1
    $ curl -X PUT -d '{"name": "John", "age": 32}' ${API_ENDPOINT}/user/1
    $ curl -X DELETE ${API_ENDPOINT}/user/1
    $ curl ${API_ENDPOINT}/user/1
    $ curl ${API_ENDPOINT}/users
    ```

## References

- https://medium.com/@kelvin_sp/building-and-testing-a-rest-api-in-golang-using-gorilla-mux-and-mysql-1f0518818ff6
- https://semaphoreci.com/community/tutorials/building-and-testing-a-rest-api-in-go-with-gorilla-mux-and-postgresql
- https://cloud.google.com/go/getting-started/using-cloud-sql
- https://github.com/lvaylet/go-chi-rest-api
- https://thenewstack.io/tutorial-deploying-a-web-application-on-google-cloud-run/
- https://github.com/GoogleCloudPlatform/golang-samples/blob/master/getting-started/bookshelf/config.go
- https://medium.com/@petomalina/connecting-to-cloud-sql-from-cloud-run-dcff2e20152a