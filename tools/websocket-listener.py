import asyncio
import websockets
import ssl
import requests
import datetime

ACCESS_TOKEN = "{access_token}"
BOX_ID = "{box_id}"

async def hello():
    ssl_context = ssl.SSLContext()
    ssl_context.check_hostname = False
    ssl_context.verify_mode = ssl.CERT_NONE
    headers = [("Origin", "https://app.misakey.com.local")]
    uri = "wss://api.misakey.com.local/boxes/{}/events/ws?access_token={}".format(BOX_ID, ACCESS_TOKEN)
    async with websockets.connect(uri, extra_headers=headers, ssl=ssl_context) as websocket:
        while True:
          resp = await websocket.recv()
          print(resp)

try:
    asyncio.get_event_loop().run_until_complete(hello())
except Exception as e:
    print("Error:", e)
