all: jsx
	go build -o websockets

jsx:
	./node_modules/.bin/jsx frontend/ web/static/js/

watch:
	./node_modules/.bin/jsx --watch frontend/ web/static/js/

serve:
	@echo "Serving on http://localhost:8080"
	./websockets

rabbit:
	docker run -d -p 15672:15672 -p 5672:5672 -e RABBITMQ_NODENAME=my-rabbit rabbitmq:3-management
