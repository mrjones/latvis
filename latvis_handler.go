package latvis_handler

import (
  "fmt"
  "http"
)

func DrawMap(response http.ResponseWriter, request *http.Request) {
     fmt.Fprintf(response, "hi");
}
