cd latitude
gomake install
cd ..

cd location
gomake install
cd ..

cd visualization
gomake install
cd ..

8g latvis_handler.go && 8g server.go && 8l server.8 && ./8.out
