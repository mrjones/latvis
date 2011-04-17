cd latitude
gomake install
cd ..

cd location
gomake install
cd ..

cd server
gomake install
cd ..

cd visualization
gomake install
cd ..

8g latvis.go && 8l latvis.8 && ./8.out --imageSize 512 && cp vis.png /var/www/vis-api-matt.png
