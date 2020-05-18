You may need to install the package at github.com/gorilla/mux

CLI command: go get -v -u github.com/gorilla/mux

To run the script, navigate to the directory and rung go run main.go

This will run the server at localhost:10000

Available endpoints are: localhost:10000/api/{github_username}/{github_repo} and localhost:10000/api/{github_username}/{github_repo}/info

The base endpoint will return all information on the available files (surface level only, ignores subdirectories), including their contents. The /info endpoint will return only the number of stars and the files (same distinction as before).

NOTE: github has placed a throttle on non-authenticated requests such as done here with a limit of 60 requests per hour. To obtain the data on each file, separate requests are run in the background. In the event that this reaches github's throttling limit, the program will substitute in a downloaded version of the repo found at github.com/google/uuid (accurate as of 5/15/2020). An alert will be surfaced in the CLI if this occurs.
