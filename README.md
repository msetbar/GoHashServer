# Hash Server in Go

This project is not scalable but it can be easily converted to a scable project. The structure of the project is flat and only using standard Go libiraies. On top of the standard libraries, two classes have been implemened to make easy to setup the routings and also a wrapper to wrap request context. There is no unit test in the project due to time but for any project, it is important to use TDD and BDD methods to implement enought tests with a good code coverage.

There are four main endpoints in this project:
    -1. POST hash: There is no input validation for incoming parameters. So the requests with no password will also be accepted.
    -2. GET hash/id: `id` should be a number and all invalid requests will be rejected with 404 status code. If a request is sent for a hash id that its hash still has not been calculated, the request will also be rejected with 404 status code.
    -3. GET stats: this API will generate a json object with the total number of requests recieved for `POST hash` along with the average processing time in microseconds for all requests for this API 
    -4. shutdown: after recieving this request, the application no longer will process any request and reply with 503 status code until all background task to calculate hashes are completed.

## How to configure port
The port number for the Hash server can be pass using an envrioment variable named `PORT` otherwise `PORT` will be set to 3000.

## How to run:
The program can be executed using VS Code or Go command. In order to use Go command, please use the following command:
    `go run server.go context.go app.go hash.go`

