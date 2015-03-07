all: jsx
	go build -o websockets

jsx:
	./node_modules/.bin/jsx frontend/ web/static/js/

watch:
	./node_modules/.bin/jsx --watch frontend/ web/static/js/

serve:
	echo "Serving on http://localhost:8000"
	( cd web/ && python -m SimpleHTTPServer )
