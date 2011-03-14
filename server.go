package main

import (
  "http"
  "./latvis_handler"
)

func main() {
     latvis_handler.DoStupidSetup()
     http.HandleFunc("/authorize", latvis_handler.Authorize);
     http.HandleFunc("/drawmap", latvis_handler.DrawMap);
     http.HandleFunc("/latestimage", latvis_handler.ServePng);
     http.ListenAndServe(":8080", nil)
}
