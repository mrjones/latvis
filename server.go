package main

import (
  "http"
  "log"
  "./latvis_handler"
)

func main() {
     latvis_handler.DoStupidSetup()
     http.HandleFunc("/authorize", latvis_handler.Authorize);
     http.HandleFunc("/drawmap", latvis_handler.DrawMap);
     http.HandleFunc("/latestimage", latvis_handler.ServePng);
     err := http.ListenAndServe(":8081", nil)
     log.Fatal(err)
}
