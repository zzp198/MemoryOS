from sanic import Sanic
from sanic import response

app = Sanic(__name__)


@app.route("/")
async def hello_world(request):
    return response.json({"hello": "world"})


@app.websocket('/stream')
async def wstest(request, ws):
    while True:
        data = await ws.recv()
        await ws.send(data)


if __name__ == "__main__":
    app.run(host='0.0.0.0', port=80, dev=True)
