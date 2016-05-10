A Websocket based Comments App
==============================

A comments application based on the official [React tutorial](http://facebook.github.io/react/docs/tutorial.html)
with a websocket server implemented with Go and RabbitMQ.

Just a work-in-progress toy project.

Usage
-----

To build everything:
```
make
```

Then run rabbit MQ in a Docker container:
```
make rabbit
```

Run the websocket service:
```
make serve
```

And run a webserver to serve the front end:
```
python -m SimpleHTTPServe
```
The application can then be acessed at http://localhost:8000
