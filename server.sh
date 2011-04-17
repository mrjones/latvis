cd location
gomake install
cd ..

cd visualization
gomake install
cd ..

8g latitude_api.go && 8g latvis_handler.go && 8g server.go && 8l server.8 && ./8.out
