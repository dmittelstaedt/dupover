# dupover

docker build -t dataport.de/dupover --build-arg http_proxy=proxy --build-arg https_proxy=proxy .
docker create -it --name dupover dataport.de/dupover:latest
docker cp dupover:/go/src/app/dupover .
