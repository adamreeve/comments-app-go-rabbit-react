Based on official react tutorial from http://facebook.github.io/react/docs/tutorial.html

While developing, run:
    ./node_modules/.bin/jsx --watch frontend/ web/static/js/
and:
    python -m SimpleHTTPServe

Then acess at http://localhost:8000

In the process of modifying this to use websockets and a server written in Go.

When not working on frontend, can just build with:
    ./node_modules/.bin/jsx frontend/ web/static/js/
