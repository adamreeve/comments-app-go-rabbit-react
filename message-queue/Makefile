CONTAINER_NAME=rabbit

build:
	go build -o rabbit-test

run:
	docker run -d -e RABBITMQ_NODENAME=my-rabbit --name ${CONTAINER_NAME} rabbitmq:3-management

start:
	docker start ${CONTAINER_NAME}

stop:
	docker stop ${CONTAINER_NAME}

ip:
	@echo "IP Address:"
	@docker inspect --format '{{ .NetworkSettings.IPAddress }}' ${CONTAINER_NAME}

management-address:
	$(eval RABBITIP=`docker inspect --format '{{ .NetworkSettings.IPAddress }}' ${CONTAINER_NAME}`)
	@echo "http://${RABBITIP}:15672"
