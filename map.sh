cd location
gomake install
cd ..

cd visualization
gomake install
cd ..

8g tokens.go  && 8g latitude_xml.go && 8g latitude_api.go && 8g latvis.go && 8l latvis.8 && ./8.out --imageSize 512 && cp vis.png /var/www/vis-api-matt.png
