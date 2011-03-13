package main

import (
  "http"
  "./latvis_handler"
)

func main() {
     http.HandleFunc("/drawmap", latvis_handler.DrawMap);
     http.ListenAndServe(":8080", nil)
}
