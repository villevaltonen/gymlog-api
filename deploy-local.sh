# deploy docker stack
docker-compose rm --stop --force -v
docker-compose -f docker-compose.yml up