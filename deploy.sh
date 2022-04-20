#!/usr/bin/env bash

docker buildx build --platform linux/amd64 -t aubipo .

# make sure to use the name of your Heroku app
docker tag aubipo registry.heroku.com/aubipo/web

# use docker push to push it to the Heroku registry
docker push registry.heroku.com/aubipo/web

# then use heroku release to activate
heroku container:release web -a aubipo