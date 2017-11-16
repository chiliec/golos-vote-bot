docker build --no-cache -t golosovalochka .
docker stop golosovalochka || true
docker rm golosovalochka || true
docker run -d -v `pwd`/db:/root/db --restart unless-stopped --name golosovalochka golosovalochka:latest .
