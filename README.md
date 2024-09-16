# datapipe

**The project is in active development and is not recommended for production usage**

**datapipe** is a highly configurable data pipeline application designed to streamline data processing through a series of customizable handlers. Each handler can retrieve, filter, transform, or output data, enabling complex data workflows to be easily managed and extended.

### Features

 - **Configurable Handlers:** Set up multiple handlers that process data in a defined sequence. Currently supported handler types include filters and HTTP handlers.
 - **Built-in filtering:** Supports complex filtering expressions within handlers, allowing for advanced data processing logic. You can compare strings and numbers using operators like `>`, `<`, `>=`, `<=`, `==`, and `!=`. The filtering engine also supports logical operators such as `&&` and `||`, as well as grouping conditions with braces for creating intricate and precise filtering rules.
 - **Flexible Workflow:** Each piece of data is processed individually by each handler, allowing for granular control and multiple result sets.
 - **Retry Logic:** Handlers can be configured with retry logic, ensuring robust and resilient data processing.
 - **Interval Execution:** Schedule your data pipeline to run at regular intervals.

### TODO
- [ ] Add a simple example handler
- [ ] Add trigger API
- [ ] Add OpenAPI doc for handlers (and triggers)
- [ ] Add cron-like scheduling
- [ ] Add a Dockerfile and publish the image to Docker Hub
- [ ] Add a CI/CD pipeline to build and test the code
- [ ] Add documentation with the full description of the config file options
- [ ] Pre-validate placeholders in the config
- [ ] Add other types of handlers with communication via gRPC, Kafka, RabbitMQ, etc.
- [ ] Add a way to save the state of the pipeline, so it can be restored after a restart
- [ ] Filter engine should precompile expressions on the start instead of compiling them on each data processing

### Example configuration:

```yaml
engine:
  disable_run_on_start: false
  interval: 24h
  log:
    level: info
    static_fields:
      app: datapipe
      env: dev

handlers:
  data-source:
    type: http
    http:
      url: http://localhost:8080/get-data
      method: GET
      headers:
        Content-Type: application/json
      query_params:
        result-interval: 24h
      timeout: 15s
      retries: 3
      retry_interval: 30s

  status-type-filter:
    type: filter
    filter:
      expression: '{{ data-source.status }} == "new"'

  example-handler:
    type: http
    http:
      url: http://localhost:8081/handle-data
      method: POST
      body: |
        {
          "data": {{ data-source.description }}
        }
      headers:
        Content-Type: application/json
      timeout: 30s

  data-sink:
    type: http
    http:
      url: http://localhost:8082/save-data
      method: POST
      body: |
        {
          "raw-data": {{ data-source.description }},
          "handler-result": {{ example-handler.result }}
        }
      headers:
        Content-Type: application/json
      timeout: 15s
      parallel_run: true
```

Workflow Explanation

 1. **data-source:** The pipeline starts with the data-source handler, which fetches data from an HTTP endpoint.
 2. **filter:** The status-type-filter processes each piece of data, filtering out irrelevant entries based on the specified criteria.
 3. **example-handler:** For each filtered data entry, the example-handler processes the data, potentially generating multiple results per input.
 4. **data-sink:** Finally, the data-sink handler consolidates the original data and handler-processed results, saving them on its side for further usage.

### HTTP handlers result format

Example of the result of an HTTP handler:


```json
{
  "result": [
    {"a": "b"},
    {"c": "d"}
  ]
}
```

---

```
Congratulations on making it this far! Here's a little surprise for your dedication:

                                                             @@@@@@@@@@@@@@@@@@@@@@@%@@@@@@@@@@@@@@@.                 
                                              ########=+#+**%:::::::::::::::::::::::::::::::::::::::@@*               
                             .#######################+=#*@=::::::::::::::::::::::::::-:-:::::::-:::::::@@             
            .*####################*###################=+@@=::::::::::::::::::=+=:::-++-::::::::::-:::::@@             
#########*+#*########*########################==========@@=::::::*++::::::::::::::::::::::::::-::::::::@@             
#######+######################==========================@@=:::::::::::::::::::::::::@@@@@::::+=+:::::::@@  @@@@@      
#############==============-==================::::::::::@@=::-:::::::::::::::::::-@@=====@%::::::::::::@@@@+====@@:   
============================-=::::::::::::::::::::::::::@@=:::::::::::::::-++::::-@@=======@@=:::-:::::@@=======@@.   
=============:::::::::::::::::::::::::::::@@@@@@@@@=.:.:@@=::-:::::::::::-:::::::-@@=========@@@@@@@@@@=========@@.   
::::::::::::::::::::::::::::::::.::::::.@@=======@@@@@%@@@=:::::::::++-::::::::::-@@============================@@:   
::::::::::::::::::::::::::::::=========-@@@@*=========@@@@=::::::::::::::::::++@@@================================@@@ 
:::::::::::::===-===========================%@@@@@@@@@==@@=::::++-:::::::::::::@@@=======  @@+===========  @@@====@@@ 
==============================****+*+*+*+*+*++*++*+%@@@@%@=::::::::::::::::::::@@@=======@@@@=======@@@==@@@@@====@@@ 
=============++*+++++++++*+++*++++++*+*+*+*++++++++*+++*@@=:::::::::::::+++::::@@@==-----=====================----@@@ 
+**+**++*+*+*+*+*+*+*+++++*++++*++*++++++++*+***********@@=::::::+++::::-:-::::@@@==-----==@@=====@@+====@@+==----@@@ 
*+*++*=+*+*+*+*+*+*+*+*++++*+*************************@@@@=::::::::-::::::::::::::@@=======@@@@@@@@@@@@@@@@+====@@:   
*+*++++++++++*********************************     %@@@@@@@@@:::::::::::::::::::::::@@+=======================@@      
*****************************#                   @@*======%@@@@@@@%#@@@@@@@@@@@@@@@@@@@%@@@@@@@@@@@@@@@@@@@@@@        
*************                                    @@*====@@-  #@-====@@              @@+====@@.  @@=====@@             
                                                 @@@@@@@       @@@@@@@                @@@@@@@.    *%@@@@*             

Your data pipeline is now cooler than ever!
```
