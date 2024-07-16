```
  go build -o bin/cmd .
  docker build -t [IMAGE NAME] .
  docker run --name [CONTAINER NAME] -p 80:8080 -v [DATA STORE]:/app/www --restart=always -e OHTTPS_TOKEN=[ohttps webhook token] -d [IMAGE NAME] 
```
